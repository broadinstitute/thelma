package kubecfg

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"path"
	"sort"
	"sync"
	"time"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/flock"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// defaultAuthInfo name of AuthInfo inside kube config file to use for authentication to non-prod clusters
const defaultAuthInfo = "default"

// ctxDelimiter delimiter to use when joining environment/cluster/chart names to build context names in kubecfg
// note that this helps guarantee uniqueness since underscores are not valid characters in env or cluster names
const ctxDelimiter = "_"

// kubectx represents a run context for a kubectl command
type kubectx struct {
	contextName string        // contextName name of the run context in the kubecfg file
	namespace   string        // namespace to run command in
	cluster     terra.Cluster // cluster the cluster this command should be executed against
}

// Kubectx represents a run context for a kubectl command
type Kubectx interface {
	// ContextName name of the run context in the kubecfg file
	ContextName() string
	// Namespace kubectl command should run against
	Namespace() string
}

func (k kubectx) ContextName() string {
	return k.contextName
}

func (k kubectx) Namespace() string {
	return k.namespace
}

// ReleaseKtx is a convenience type that bundles a terra.Release with its associated Kubectx
type ReleaseKtx struct {
	Release terra.Release
	Kubectx Kubectx
}

// Kubeconfig manages entries in a `kubectl` config file (traditionally ~/.kube/config) for Terra environments & releases.
// It creates context entries for environments and releases.
// It works like `gcloud container clusters get-credentials`, except:
//   - Users don't have to specify a project or location (because thelma already knows where clusters live)
//   - Context entries are named in a more user-friendly fashion. For example, the context for the "alpha" environment
//     is called `alpha`, instead of `gke_broad-dsde-alpha_us-central1-a_terra-alpha`. This makes it possible to quickly run
//     a kubectl command against the alpha environment using `kubectl -c alpha` (no need to specify a namespace).
//
// Read more about kubectl contexts here:
// https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#define-clusters-users-and-contexts
//
// Q: Why not just shell out to gcloud?
// Because `gcloud` is a big ol' Python app. Thelma is designed to run on developer laptops and depending on users
// having the correct version of gcloud and Python installed is more brittle than just generating GKE credentials ourselves.
type Kubeconfig interface {
	// ConfigFile path to the .kubecfg file where context entries are generated
	ConfigFile() string
	// ForRelease returns the name of the kubectx to use for executing commands against this release
	ForRelease(terra.Release) (Kubectx, error)
	// ForReleases returns a kubectx for each given release
	ForReleases(releases ...terra.Release) ([]ReleaseKtx, error)
	// ForEnvironment returns all kubectxs for all releases in an environment
	ForEnvironment(env terra.Environment) ([]Kubectx, error)
	// ForCluster returns the name of the kubectx to use for executing commands against the cluster (without any awareness of environment/releases)
	ForCluster(cluster terra.Cluster) (Kubectx, error)
}

// New constructs a Kubeconfig
func New(file string, gkeClient *container.ClusterManagerClient, tokenSource oauth2.TokenSource) Kubeconfig {
	lockfile := path.Join(path.Dir(file), "."+path.Base(file)+".lk")

	return &kubeconfig{
		file:        file,
		gkeClient:   gkeClient,
		tokenSource: tokenSource,
		locker: flock.NewLocker(lockfile, func(options *flock.Options) {
			options.Timeout = 30 * time.Second
			options.RetryInterval = 100 * time.Millisecond
		}),
		writtenCtxs: make(map[string]struct{}),
	}
}

type kubeconfig struct {
	file        string                          // path to kubeconfig file we should write auth creds to (eg. ~/.thelma/config/kubeconfig)
	gkeClient   *container.ClusterManagerClient // google container engine / kubernetes engine client
	tokenSource oauth2.TokenSource              // token source to be used when adding auth token to kubeconfig file
	locker      flock.Locker                    // file lock for preventing concurrent kubeconfig updates from stomping on each other
	writtenCtxs map[string]struct{}             // cache for previously-written kubectxs
	mutex       sync.Mutex                      // mutex for safe read/writing to writtenContexts
}

func (c *kubeconfig) ForRelease(release terra.Release) (Kubectx, error) {
	return c.addRelease(release)
}

func (c *kubeconfig) ForReleases(releases ...terra.Release) ([]ReleaseKtx, error) {
	var result []ReleaseKtx
	for _, r := range releases {
		ktx, err := c.ForRelease(r)
		if err != nil {
			return nil, err
		}
		result = append(result, ReleaseKtx{
			Release: r,
			Kubectx: ktx,
		})
	}
	return result, nil
}

func (c *kubeconfig) ForEnvironment(env terra.Environment) ([]Kubectx, error) {
	return c.addAllReleases(env)
}

func (c *kubeconfig) ForCluster(cluster terra.Cluster) (Kubectx, error) {
	return c.addCluster(cluster)
}

func (c *kubeconfig) ConfigFile() string {
	return c.file
}

// addAllReleases updates the kubecfg to include a context for all releases in an environment.
//
// This means adding the environment's default context (see addEnvironmentDefault), and, for any
// release outside the environment's default cluster, a separate context keyed by the releases
// Argo application name, which is globally unique.
//
// For example, when called for Terra's alpha environment, addAllReleases will add a default context
// called "alpha" as well as a context called "datarepo-alpha", which points at DataRepo's alpha cluster
// and the "terra-alpha" namespace.
func (c *kubeconfig) addAllReleases(env terra.Environment) ([]Kubectx, error) {
	var kubectxts []Kubectx

	// add environment's default context
	defaultCtx, err := c.addEnvironmentDefault(env)
	if err != nil {
		return nil, err
	}
	kubectxts = append(kubectxts, defaultCtx)

	// add context for any releases that aren't in the environment's default cluster
	for _, release := range env.Releases() {
		if release.Cluster().Name() != env.DefaultCluster().Name() {
			_kubectx, err := c.addRelease(release)
			if err != nil {
				return nil, err
			}
			kubectxts = append(kubectxts, _kubectx)
		}
	}

	// sort by context name for predictability and easy testing
	sort.Slice(kubectxts, func(i, j int) bool {
		return kubectxts[i].ContextName() < kubectxts[j].ContextName()
	})

	return kubectxts, nil
}

// addEnvironmentDefault updates the kubecfg to include a context for the environment, pointing at the environment's
// namespace and default cluster. For example, if called for Terra's alpha environment, this will add a context to
// the kubeconfig that is called "alpha", points at the "terra-alpha" namespace, and is configured to point at the
// terra-alpha cluster in the broad-dsde-alpha project.
//
// Note that this does NOT generate a context for releases that live outside the environent's default cluster.
// (eg. "datarepo"). To generate a context for those releases as well, call addAllReleases() instead.
func (c *kubeconfig) addEnvironmentDefault(env terra.Environment) (Kubectx, error) {
	_kubectx := kubectxForEnvironment(env)
	err := c.writeContextIfNeeded(_kubectx)
	if err != nil {
		return kubectx{}, err
	}
	return _kubectx, nil
}

// addRelease adds a context for this release to the kubecfg file.
//
// If this is an application release deployed to an environment's default cluster, this will add the environment
// default context (see addEnvironmentDefault).
//
// Else, this will add context for this release that is keyed by the release's Argo application name and points
// at the release's target cluster and namespace.
func (c *kubeconfig) addRelease(release terra.Release) (Kubectx, error) {
	_kubectx := kubectxForRelease(release)
	err := c.writeContextIfNeeded(_kubectx)
	if err != nil {
		return kubectx{}, err
	}
	return _kubectx, nil
}

// addCluster updates the kubecfg to include a context for given cluster, pointing at the cluster's default
// namespace.
//
// For example, if called for Terra's alpha cluster, this will add a context to
// the kubeconfig that is called "cluster_terra-alpha", uses the default namespace, and is (of course) pointing
// at the terra-alpha cluster in the broad-dsde-alpha project.
func (c *kubeconfig) addCluster(cluster terra.Cluster) (Kubectx, error) {
	_kubectx := kubectxForCluster(cluster)
	err := c.writeContextIfNeeded(_kubectx)
	if err != nil {
		return kubectx{}, err
	}
	return _kubectx, nil
}

// writeContextIfNeeded writes a context with the given name, namespace, and target cluster to the kubeconfig file,
// unless the same context has already been written at least once by this kubecfg instance.
// (saves Thelma from making a Google Cloud api call for every `kubectl` command it runs)
func (c *kubeconfig) writeContextIfNeeded(ctx kubectx) error {
	c.mutex.Lock()
	_, exists := c.writtenCtxs[ctx.contextName]
	c.mutex.Unlock()

	if exists {
		// this context has already been written once by this kubecfg, no need to write again.
		return nil
	}

	err := c.writeContext(ctx)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.writtenCtxs[ctx.contextName] = struct{}{}
	c.mutex.Unlock()

	return nil
}

// writeContext writes a context with the given name, namespace and target cluster to the kubeconfig file
// a file lock is used to prevent concurrent updates from stomping on one another
func (c *kubeconfig) writeContext(ctx kubectx) error {
	log.Debug().
		Str("context", ctx.contextName).
		Str("namespace", ctx.namespace).
		Str("cluster", ctx.cluster.Name()).
		Msgf("Generating %s entry for %s", c.file, ctx.contextName)

	err := c.locker.WithLock(func() error {
		return c.writeContextUnsafe(ctx)
	})
	if err != nil {
		return errors.Errorf("error generating kubectx for cluster %s: %v", ctx.cluster.Name(), err)
	}
	return nil
}

// writeContextUnsafe writes a context with the given name, namespace and target cluster to the kubeconfig file
// (it does not synchronize write access, hence "unsafe")
func (c *kubeconfig) writeContextUnsafe(ctx kubectx) error {
	cfg, err := c.readKubecfg()
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = &clientcmdapi.Config{}
	}

	cluster := ctx.cluster

	clusterData, err := c.gkeClient.GetCluster(context.Background(), &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s",
			cluster.Project(),
			cluster.Location(),
			cluster.Name()),
	})

	if err != nil {
		return err
	}

	// Add cluster CA certificate
	caCert, err := base64.StdEncoding.DecodeString(clusterData.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return err
	}

	// Add cluster definition to kubeconfig
	if len(cfg.Clusters) == 0 {
		cfg.Clusters = make(map[string]*clientcmdapi.Cluster)
	}
	cfg.Clusters[cluster.Name()] = &clientcmdapi.Cluster{
		Server:                   cluster.Address(),
		CertificateAuthorityData: caCert,
	}

	if len(cfg.Contexts) == 0 {
		cfg.Contexts = make(map[string]*clientcmdapi.Context)
	}
	cfg.Contexts[ctx.contextName] = &clientcmdapi.Context{
		Cluster:   cluster.Name(),
		Namespace: ctx.namespace,
		AuthInfo:  defaultAuthInfo,
	}

	if len(cfg.AuthInfos) == 0 {
		cfg.AuthInfos = make(map[string]*clientcmdapi.AuthInfo)
	}

	token, err := c.tokenSource.Token()
	if err != nil {
		return err
	}

	// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuration
	cfg.AuthInfos[defaultAuthInfo] = &clientcmdapi.AuthInfo{
		// TODO we can use an exec command here to run thelma and print out the auth token.
		// This would make kubecfg entries stable/re-usable, so we only need to generate them once.
		// Eg. 	Exec: &clientcmdapi.ExecConfig{
		//			Command: "thelma",
		//			Args:    []string{"auth", "gcp", "--access-token", "--cluster", cluster.Name(), "--echo"},
		//		},
		Token: token.AccessToken,
	}

	return clientcmd.WriteToFile(*cfg, c.file)
}

// read kube config file, returning nil if it does not exist
func (c *kubeconfig) readKubecfg() (*clientcmdapi.Config, error) {
	exists, err := utils.FileExists(c.file)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return clientcmd.LoadFromFile(c.file)
}

func kubectxForEnvironment(env terra.Environment) kubectx {
	return kubectx{
		contextName: env.Name(),
		cluster:     env.DefaultCluster(),
		namespace:   env.Namespace(),
	}
}

func kubectxForRelease(release terra.Release) kubectx {
	return kubectx{
		contextName: contextNameForRelease(release),
		cluster:     release.Cluster(),
		namespace:   release.Namespace(),
	}
}

func kubectxForCluster(cluster terra.Cluster) kubectx {
	return kubectx{
		contextName: contextNameForCluster(cluster),
		cluster:     cluster,
	}
}

func contextNameForCluster(cluster terra.Cluster) string {
	// prefix with cluster to avoid collisions with environment context names
	return "cluster" + ctxDelimiter + cluster.Name()
}

// contextNameForRelease computes name of the kubecfg context for this release.
// If the release is deployed to an environment's default cluster, then use the
// environment name. ("alpha", "staging", "fiab-choover-funky-squirrel")
// If the release is a cluster release, or if the release is deployed to a different cluster than the environment's
// default, use the ArgoCD application name, which is globally unique (eg. "datarepo_staging")
func contextNameForRelease(release terra.Release) string {
	if release.IsClusterRelease() {
		return release.Cluster().Name() + ctxDelimiter + release.Name()
	}
	appRelease, ok := release.(terra.AppRelease)
	if !ok {
		panic(errors.Errorf("failed to cast to AppRelease: %v", appRelease))
	}
	if appRelease.Cluster().Name() != appRelease.Environment().DefaultCluster().Name() {
		return appRelease.Environment().Name() + ctxDelimiter + release.Name()
	}
	return appRelease.Environment().Name()
}

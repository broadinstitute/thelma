package kubectl

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/flock"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"path"
	"time"
)

// kubectx represents a run context for a kubectl command
type kubectx struct {
	contextName string        // contextName name of the run context in kubecfg
	namespace   string        // namespace to run command in
	cluster     terra.Cluster // cluster the cluster this command should be executed against
}

// newKubeConfig constructs a kubeconfig
func newKubeConfig(file string, gkeClient *container.ClusterManagerClient, tokenSource oauth2.TokenSource) *kubeconfig {
	lockfile := path.Join(path.Dir(file), "."+path.Base(file)+".lk")

	return &kubeconfig{
		file:        file,
		gkeClient:   gkeClient,
		tokenSource: tokenSource,
		locker: flock.NewLocker(lockfile, func(options *flock.Options) {
			options.Timeout = 30 * time.Second
			options.RetryInterval = 100 * time.Millisecond
		}),
	}
}

// kubeconfig manages entries in a `kubectl` config file (traditionally ~/.kube/config) for Terra environments & releases.
// It creates context entries for environments and releases.
// It works like `gcloud container clusters get-credentials`, except:
//   - Users don't have to specify a project or location (because thelma already knows where clusters live)
//   - Context entries are named in a more user-friendly fashion. For example, the context for the "alpha" environment
//     is called `alpha`, instead of `gke_broad-dsde-alpha_us-central1-a_terra-alpha`. This makes it possible to quickly run
//     a kubectl command against the alpha environment using `kubectl -c alpha` (no need to specify a namespace).
//
// Read more about kubectl contexts here:
// https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#define-clusters-users-and-contexts
type kubeconfig struct {
	file        string                          // path to kubeconfig file (eg. ~/.thelma/config/kubeconfig)
	gkeClient   *container.ClusterManagerClient // google container engine / kubernetes engine client
	tokenSource oauth2.TokenSource              // token source to be used when adding auth token to kubeconfig file
	locker      flock.Locker                    // file lock for preventing concurrent kubeconfig updates from stomping on each other
}

// addEnvironmentDefault updates the kubecfg to include a context for the environment, pointing at the environment's
// namespace and default cluster. For example, if called for Terra's alpha environment, this will add a context to
// the kubeconfig that is called "alpha", points at the "terra-alpha" namespace, and is configured to point at the
// terra-alpha cluster in the broad-dsde-alpha project.
//
// Note that this does NOT generate a context for releases that live outside the environent's default cluster.
// (eg. "datarepo"). To generate a context for those releases as well, call addAllReleases() instead.
func (c *kubeconfig) addEnvironmentDefault(env terra.Environment) (kubectx, error) {
	_kubectx := c.kubectxForEnvironment(env)
	err := c.writeContext(_kubectx)
	if err != nil {
		return kubectx{}, err
	}
	return _kubectx, nil
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
func (c *kubeconfig) addAllReleases(env terra.Environment) ([]kubectx, error) {
	var kubectxts []kubectx

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

	return kubectxts, nil
}

// addRelease adds a context for this release to the kubecfg file.
//
// If this is an application release deployed to an environment's default cluster, this will add the environment
// default context (see addEnvironmentDefault).
//
// Else, this will add context for this release that is keyed by the release's Argo application name and points
// at the release's target cluster and namespace.
func (c *kubeconfig) addRelease(release terra.Release) (kubectx, error) {
	_kubectx := c.kubectxForRelease(release)
	err := c.writeContext(_kubectx)
	if err != nil {
		return kubectx{}, err
	}
	return _kubectx, nil
}

// writeContext writes a context with the given name, namespace and target cluster to the kubeconfig file
// a file lock is used to prevent concurrent updates from stomping on one another
func (c *kubeconfig) writeContext(ctx kubectx) error {
	log.Debug().
		Str("context", ctx.contextName).
		Str("namespace", ctx.namespace).
		Str("cluster", ctx.cluster.Name()).
		Msgf("Generating %s entry for %s", c.file, ctx.contextName)

	return c.locker.WithLock(func() error {
		return c.writeContextUnsafe(ctx.contextName, ctx.namespace, ctx.cluster)
	})
}

// writeContextUnsafe writes a context with the given name, namespace and target cluster to the kubeconfig file
// (it does not synchronize write access, hence "unsafe")
func (c *kubeconfig) writeContextUnsafe(contextName string, namespace string, cluster terra.Cluster) error {
	cfg, err := c.readKubecfg()
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = &clientcmdapi.Config{}
	}

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
	cfg.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:   cluster.Name(),
		Namespace: namespace,
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

func (c *kubeconfig) kubectxForEnvironment(env terra.Environment) kubectx {
	return kubectx{
		contextName: env.Name(),
		cluster:     env.DefaultCluster(),
		namespace:   env.Namespace(),
	}
}

func (c *kubeconfig) kubectxForRelease(release terra.Release) kubectx {
	return kubectx{
		contextName: c.contextNameForRelease(release),
		cluster:     release.Cluster(),
		namespace:   release.Namespace(),
	}
}

// contextForRelease computes name of the kubecfg context for this release.
// If the release is deployed to an environment's default cluster, then use the
// environment name. ("alpha", "staging", "fiab-choover-funky-squirrel")
// If the release is a cluster release, or if the release is deployed to a different cluster than the environment's
// default, use the ArgoCD application name, which is globally unique (eg. "datarepo-staging")
func (c *kubeconfig) contextNameForRelease(release terra.Release) string {
	if release.IsClusterRelease() {
		return argocd.ApplicationName(release)
	}
	appRelease, ok := release.(terra.AppRelease)
	if !ok {
		panic(fmt.Errorf("failed to cast to AppRelease: %v", appRelease))
	}
	if appRelease.Cluster().Name() != appRelease.Environment().DefaultCluster().Name() {
		return argocd.ApplicationName(release)
	}
	return appRelease.Environment().Name()
}

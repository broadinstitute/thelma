package kubernetes

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path"
)

// kubeConfigName name of the kube config file generated / updated by Thelma, stored in ~/.thelma/config/
const kubeConfigName = "kubecfg"

// Clients factory interface for K8s tools & clients
type Clients interface {
	// ForRelease returns an API client authenticated against the release's cluster
	ForRelease(release terra.Release) (k8s.Interface, error)
	// ForKubectx returns an API client authenticated against the cluster referred to by the given Kubeconfig
	ForKubectx(kubectx kubecfg.Kubectx) (k8s.Interface, error)
	// Kubectl returns a new Kubectl
	Kubectl() (kubectl.Kubectl, error)
	// Kubecfg returns the Kubecfg instance used by this Clients
	Kubecfg() (kubecfg.Kubeconfig, error)
}

func New(thelmaRoot root.Root, shellRunner shell.Runner, googleClients google.Clients) Clients {
	return &clients{
		kubecfg:     newLazyKubecfg(thelmaRoot, googleClients),
		shellRunner: shellRunner,
	}
}

type clients struct {
	kubecfg     lazy.LazyE[kubecfg.Kubeconfig]
	shellRunner shell.Runner
}

func (k *clients) ForRelease(release terra.Release) (k8s.Interface, error) {
	_kubecfg, err := k.kubecfg.Get()
	if err != nil {
		return nil, err
	}
	kubectx, err := _kubecfg.ForRelease(release)
	if err != nil {
		return nil, err
	}
	return k.ForKubectx(kubectx)
}

func (k *clients) ForKubectx(kubectx kubecfg.Kubectx) (k8s.Interface, error) {
	_kubecfg, err := k.kubecfg.Get()
	if err != nil {
		return nil, err
	}

	parsedKubecfg, err := clientcmd.LoadFromFile(_kubecfg.ConfigFile())
	if err != nil {
		return nil, err
	}

	var overrides clientcmd.ConfigOverrides
	overrides.CurrentContext = kubectx.ContextName()

	clientConfig := clientcmd.NewDefaultClientConfig(*parsedKubecfg, &overrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return k8s.NewForConfig(restConfig)
}

func (k *clients) Kubectl() (kubectl.Kubectl, error) {
	_kubecfg, err := k.kubecfg.Get()
	if err != nil {
		return nil, err
	}
	return kubectl.New(k.shellRunner, _kubecfg), nil
}

func (k *clients) Kubecfg() (kubecfg.Kubeconfig, error) {
	return k.kubecfg.Get()
}

func newLazyKubecfg(thelmaRoot root.Root, googleClients google.Clients) lazy.LazyE[kubecfg.Kubeconfig] {
	return lazy.NewLazyE[kubecfg.Kubeconfig](func() (kubecfg.Kubeconfig, error) {
		return buildKubecfg(thelmaRoot, googleClients)
	})
}

func buildKubecfg(thelmaRoot root.Root, googleClients google.Clients) (kubecfg.Kubeconfig, error) {
	kubecfgFile := path.Join(thelmaRoot.ConfigDir(), kubeConfigName)

	clusterManagerClient, err := googleClients.ClusterManager()
	if err != nil {
		return nil, err
	}
	tokenSource, err := googleClients.TokenSource()
	if err != nil {
		return nil, err
	}

	return kubecfg.New(kubecfgFile, clusterManagerClient, tokenSource), nil
}

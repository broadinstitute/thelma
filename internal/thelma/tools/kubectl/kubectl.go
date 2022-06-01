package kubectl

import (
	container "cloud.google.com/go/container/apiv1"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"path"
)

// kubeConfigName name of the kube config file generated / updated by Thelma, stored in ~/.thelma/config/
const kubeConfigName = "kubecfg"

// defaultAuthInfo name of AuthInfo inside kube config to use for authentication to non-prod clusters
const defaultAuthInfo = "default"

// prog name of the kubectl binary we're executing
const prog = `kubectl`

// kubeConfigEnvVar env var to use to pass kube config file path to `kubectl`
const kubeConfigEnvVar = "KUBECONFIG"

// https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuration

// Kubectl is a golang interface for executing `kubectl` commands
type Kubectl interface {
	// ShutDown downscales all statefulsets and deployments in the environment to 0 replicas
	ShutDown(env terra.Environment) error
	// DeletePVCs will delete all persistent volume claims in the environment
	DeletePVCs(env terra.Environment) error
}

func NewKubectl(shellRunner shell.Runner, thelmaRoot root.Root, tokenSource oauth2.TokenSource, gkeClient *container.ClusterManagerClient) (Kubectl, error) {
	configFile := path.Join(thelmaRoot.ConfigDir(), kubeConfigName)

	return &kubectl{
		configFile:  configFile,
		shellRunner: shellRunner,
		kubeconfig:  newKubeConfig(configFile, gkeClient, tokenSource),
	}, nil
}

// implements the Kubectl interface
type kubectl struct {
	configFile  string
	shellRunner shell.Runner
	kubeconfig  *kubeconfig
}

func (k *kubectl) ShutDown(env terra.Environment) error {
	log.Info().Msgf("Downscaling all deployments in %s to 0 replicas", env.Name())
	if err := k.runForEnv(env, []string{"scale", "--replicas=0", "deployment", "--all"}); err != nil {
		return err
	}
	log.Info().Msgf("Downscaling all statefulsets in %s to 0 replicas", env.Name())
	if err := k.runForEnv(env, []string{"scale", "--replicas=0", "statefulset", "--all"}); err != nil {
		return err
	}
	return nil
}

func (k *kubectl) DeletePVCs(env terra.Environment) error {
	if env.Lifecycle() != terra.Dynamic {
		// Guard against kabooming data in long-lived static environments (such as, for example, prod)
		return fmt.Errorf("DeletePVCs can only be called for dynamic environments")
	}
	log.Info().Msgf("Deleting all PVCs in %s", env.Name())
	return k.runForEnv(env, []string{"delete", "persistentvolumeclaims", "--all", "--wait=true"})
}

// runForEnv will run a kubectl command for all of an environment's contexts
func (k *kubectl) runForEnv(env terra.Environment, args []string) error {
	kubectxs, err := k.kubeconfig.addAllReleases(env)
	if err != nil {
		return err
	}

	for _, _kubectx := range kubectxs {
		if err := k.runCmd(_kubectx, args); err != nil {
			return err
		}
	}

	return nil
}

func (k *kubectl) runCmd(_kubectx kubectx, args []string) error {
	kargs := []string{"--context", _kubectx.contextName, "--namespace", _kubectx.namespace}
	kargs = append(kargs, args...)

	return k.shellRunner.Run(shell.Command{
		Prog: prog,
		Args: kargs,
		Env: []string{
			fmt.Sprintf("%s=%s", kubeConfigEnvVar, k.configFile),
		},
	})
}

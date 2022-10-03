package kubectl

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	// DeleteNamespace will delete the environment's namespace
	DeleteNamespace(env terra.Environment) error
	// CreateNamespace will create the environment's namespace
	CreateNamespace(env terra.Environment) error
	// PortForward runs `kubectl port-forward` and returns the forwarding local port, a callback to stop forwarding, and
	// a possible error if the command failed.
	// The targetResource should be of the form `[pods|deployment|replicaset|service]/<name>`, like `service/sam-postgres-service`.
	PortForward(targetRelease terra.Release, targetResource string, targetPort int) (int, func() error, error)
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

func (k *kubectl) CreateNamespace(env terra.Environment) error {
	log.Info().Msgf("Creating environment namespace: %s", env.Namespace())
	return k.runForEnv(env, []string{"create", "namespace", env.Namespace()})
}

func (k *kubectl) DeleteNamespace(env terra.Environment) error {
	if env.Lifecycle() != terra.Dynamic {
		// Guard against kabooming data in long-lived static environments (such as, for example, prod)
		return fmt.Errorf("DeleteNamespace can only be called for dynamic environments")
	}
	log.Info().Msgf("Deleting environment namespace: %s", env.Namespace())
	// Note: We supply --ignore-not-found because sometimes BEE creation fails before the namespace is created.
	// This allows `thelma bee delete` to run successfully even when that happens.
	return k.runForEnv(env, []string{"delete", "namespace", "--ignore-not-found", env.Namespace()})
}

func (k *kubectl) PortForward(targetRelease terra.Release, targetResource string, targetPort int) (int, func() error, error) {
	log.Debug().Msgf("Port-forwarding to %s on port %d (in %s cluster's %s namespace)", targetResource, targetPort, targetRelease.ClusterName(), targetRelease.Namespace())
	kubectx, err := k.kubeconfig.addRelease(targetRelease)
	if err != nil {
		return 0, nil, err
	}
	output := &strings.Builder{}
	cmd := k.makeCmd(kubectx, []string{"port-forward", targetResource, fmt.Sprintf(":%d", targetPort)})
	subprocess := k.shellRunner.PrepareSubprocess(cmd, func(options *shell.RunOptions) {
		options.Stdout = output
		options.Stderr = output
	})
	err = subprocess.Start()
	if err != nil {
		return 0, nil, err
	}
	portParserCtx, cancelPortParser := context.WithCancel(context.Background())
	defer cancelPortParser()
	portChannel := make(chan int)
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			select {
			default:
				port := parsePortFromPortForwardOutput(output.String())
				if port > 0 {
					portChannel <- port
					return
				}
			case <-portParserCtx.Done():
				return
			}
		}
	}()
	select {
	case port := <-portChannel:
		return port, func() error {
			log.Debug().Msgf("Stopping port-forwarding to %s", targetResource)
			return subprocess.Stop()
		}, nil
	case <-time.After(10 * time.Second):
		_ = subprocess.Stop()
		return 0, nil, fmt.Errorf("kubectl port-forward output didn't yield a local port within 10 seconds, output: \n%s", output.String())
	}
}

var portRegex = regexp.MustCompile(`\d+\.\d+\.\d+\.\d+:(\d+)`)

func parsePortFromPortForwardOutput(output string) int {
	if matched := portRegex.FindStringSubmatch(output); len(matched) > 1 {
		// matched will be like ["127.0.0.1:1234" "1234"]
		port, err := strconv.Atoi(matched[1])
		if err != nil {
			return 0
		}
		return port
	}
	return 0
}

// runForEnv will run a kubectl command for all of an environment's contexts
func (k *kubectl) runForEnv(env terra.Environment, args []string) error {
	kubectxs, err := k.kubeconfig.addAllReleases(env)
	if err != nil {
		return err
	}

	for _, _kubectx := range kubectxs {
		if err := k.shellRunner.Run(k.makeCmd(_kubectx, args)); err != nil {
			return err
		}
	}

	return nil
}

func (k *kubectl) makeCmd(_kubectx kubectx, args []string) shell.Command {
	kargs := []string{"--context", _kubectx.contextName, "--namespace", _kubectx.namespace}
	kargs = append(kargs, args...)

	return shell.Command{
		Prog: prog,
		Args: kargs,
		Env: []string{
			fmt.Sprintf("%s=%s", kubeConfigEnvVar, k.configFile),
		},
	}
}

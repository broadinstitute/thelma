package kubectl

import (
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// prog name of the kubectl binary we're executing
const prog = `kubectl`

// kubeConfigEnvVar env var to use to pass kube config file path to `kubectl`
const kubeConfigEnvVar = "KUBECONFIG"

type LogsOptions struct {
	// Writer optional callback that should return a writer where a given container's logs should be streamed
	// If nil, logs are streamed to stdout
	Writer io.Writer
	// ContainerName optional container name to specify with -c
	ContainerName string
	// MaxLines maximum number of log lines to retrieve
	MaxLines int
}

type LogsOption func(options *LogsOptions)

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
	// Logs runs `kubectl logs` against the given Kubectx with given parameters
	Logs(ktx kubecfg.Kubectx, podSelector map[string]string, option ...LogsOption) error
	// Exec runs `kubectl exec` for the given Kubectx and pod with given parameters
	Exec(ktx kubecfg.Kubectx, container Container, command []string, opts ...shell.RunOption) error
	// ExecInteractive runs `kubectl exec` with stdin/stdout/stderr connected to local OS stdout/stdin/stderr
	ExecInteractive(ktx kubecfg.Kubectx, container Container, command []string) error
	// PortForward runs `kubectl port-forward` and returns the forwarding local port, a callback to stop forwarding, and
	// a possible error if the command failed.
	// The targetResource should be of the form `[pods|deployment|replicaset|service]/<name>`, like `service/sam-postgres-service`.
	PortForward(targetRelease terra.Release, targetResource string, targetPort int) (int, func() error, error)
}

func New(shellRunner shell.Runner, kubeconfig kubecfg.Kubeconfig) Kubectl {
	return &kubectl{
		shellRunner: shellRunner,
		kubeconfig:  kubeconfig,
	}
}

// implements the Kubectl interface
type kubectl struct {
	shellRunner shell.Runner
	kubeconfig  kubecfg.Kubeconfig
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
	kubectx, err := k.kubeconfig.ForRelease(targetRelease)
	if err != nil {
		return 0, nil, err
	}
	output := &strings.Builder{}
	cmd := k.makeCmd(kubectx, kubectx.Namespace(), []string{"port-forward", targetResource, fmt.Sprintf(":%d", targetPort)})
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

func (k *kubectl) Logs(ktx kubecfg.Kubectx, podSelector map[string]string, opts ...LogsOption) error {
	options := LogsOptions{
		Writer:        os.Stdout,
		ContainerName: "",
		MaxLines:      0,
	}
	for _, opt := range opts {
		opt(&options)
	}

	args := []string{
		"logs",
		"--selector",
		joinSelectorLabels(podSelector),
	}
	if options.ContainerName != "" {
		args = append(args, "--container", options.ContainerName)
	}
	if options.MaxLines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", options.MaxLines))
	}
	return k.runForKubectx(ktx, args, func(runOpts *shell.RunOptions) {
		runOpts.Stdout = options.Writer
		// don't send container logs to thelma's logging system -- it's noisy AF and is a significant perf hit
		runOpts.LogStdout = false
	})
}

type Container struct {
	Pod       string
	Namespace string
	Name      string
}

func (k *kubectl) Exec(ktx kubecfg.Kubectx, container Container, command []string, shellOpts ...shell.RunOption) error {
	args := []string{"exec", container.Pod, "-c", container.Name, "--"}
	args = append(args, command...)
	return k.runForKubectxWithNamespace(ktx, container.Namespace, args, shellOpts...)
}

func (k *kubectl) ExecInteractive(ktx kubecfg.Kubectx, container Container, command []string) error {
	args := []string{"exec", "-it", container.Pod, "-c", container.Name, "--"}
	args = append(args, command...)
	return k.runForKubectxWithNamespace(ktx, container.Namespace, args, func(options *shell.RunOptions) {
		options.Stdin = os.Stdin
		options.Stdout = os.Stdout
		options.Stderr = os.Stderr
	})
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
	kubectxs, err := k.kubeconfig.ForEnvironment(env)
	if err != nil {
		return err
	}

	for _, _kubectx := range kubectxs {
		if err := k.runForKubectx(_kubectx, args); err != nil {
			return err
		}
	}

	return nil
}

func (k *kubectl) runForKubectx(kubectx kubecfg.Kubectx, args []string, opts ...shell.RunOption) error {
	return k.runForKubectxWithNamespace(kubectx, kubectx.Namespace(), args, opts...)
}

func (k *kubectl) runForKubectxWithNamespace(kubectx kubecfg.Kubectx, namespace string, args []string, opts ...shell.RunOption) error {
	return k.shellRunner.Run(k.makeCmd(kubectx, namespace, args), opts...)
}

func (k *kubectl) makeCmd(kubectx kubecfg.Kubectx, namespace string, args []string) shell.Command {
	kargs := []string{"--context", kubectx.ContextName(), "--namespace", namespace}
	kargs = append(kargs, args...)

	return shell.Command{
		Prog: prog,
		Args: kargs,
		Env: []string{
			fmt.Sprintf("%s=%s", kubeConfigEnvVar, k.kubeconfig.ConfigFile()),
		},
	}
}

func joinSelectorLabels(labels map[string]string) string {
	var pairs []string

	for k, v := range labels {
		pairs = append(pairs, k+"="+v)
	}

	sort.Strings(pairs)

	return strings.Join(pairs, ",")
}

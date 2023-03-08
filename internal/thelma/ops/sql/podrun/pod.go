package podrun

import (
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

// Pod represents a pod launched in the cluster by Runner
type Pod interface {
	// Exec execute a command inside the pod's client container
	Exec(cmd []string) error
	// ExecInteractive executes a command inside the pod's client container, with stdin/stdout/stderr
	// connected to current process's stdin/stdout/stderr
	ExecInteractive(cmd []string) error
	// Delete delete pod
	Delete() error
	// Close alias for Delete
	Close() error
}

type pod struct {
	name          string
	envSecret     string
	scriptsSecret string
	kubectl       kubectl.Kubectl
	kubectx       kubecfg.Kubectx
	client        kubernetes.Interface
}

func (p *pod) Exec(cmd []string) error {
	return p.kubectl.Exec(p.kubectx, kubectl.Container{
		Pod:       p.name,
		Namespace: thelmaWorkloadsNamespace,
		Name:      clientContainerName,
	}, cmd)
}

func (p *pod) ExecInteractive(cmd []string) error {
	start := time.Now()

	err := p.kubectl.ExecInteractive(p.kubectx, kubectl.Container{
		Pod:       p.name,
		Namespace: thelmaWorkloadsNamespace,
		Name:      clientContainerName,
	}, cmd)

	elapsed := time.Now().Sub(start)

	if elapsed > time.Minute && errors.Is(err, &shell.ExitError{}) {
		log.Debug().Err(err).Msgf("ignoring exit error from interactive shell")
	}

	return err
}

func (p *pod) Delete() error {
	var err error

	log.Info().Msgf("Deleting pod: %s", p.name)
	if err = p.client.CoreV1().Pods(thelmaWorkloadsNamespace).Delete(context.Background(), p.name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("error deleting pod: %v", err)
	}

	log.Info().Msgf("Deleting scripts secret: %s", p.scriptsSecret)
	if err = p.client.CoreV1().Secrets(thelmaWorkloadsNamespace).Delete(context.Background(), p.scriptsSecret, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("error deleting cm: %v", err)
	}

	log.Info().Msgf("Deleting env secret: %s", p.envSecret)
	if err = p.client.CoreV1().Secrets(thelmaWorkloadsNamespace).Delete(context.Background(), p.envSecret, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("error deleting secret: %v", err)
	}

	return nil
}

func (p *pod) Close() error {
	return p.Delete()
}

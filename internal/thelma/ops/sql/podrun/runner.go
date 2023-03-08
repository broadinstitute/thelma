package podrun

import (
	"context"
	"fmt"
	k8s "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"strconv"
	"time"
)

const thelmaWorkloadsNamespace = "thelma-workloads"
const clientContainerName = "sqlclient"

// how long we expect one of these pods to realistically be in use
const maxPodLifetimeSeconds = 3600 * 10
const podReadinessTimeout = 5 * time.Minute

type Runner interface {
	Create(Spec) (Pod, error)
	Cleanup() error
}

func New(connection api.Connection, clients k8s.Clients) (Runner, error) {
	kcfg, err := clients.Kubecfg()
	if err != nil {
		return nil, err
	}
	kctx, err := kcfg.ForCluster(connection.Options.ProxyCluster)
	if err != nil {
		return nil, err
	}
	_kubectl, err := clients.Kubectl()
	if err != nil {
		return nil, err
	}
	client, err := clients.ForKubectx(kctx)
	if err != nil {
		return nil, err
	}
	return &runner{
		connection: connection,
		kubectl:    _kubectl,
		client:     client,
		kubectx:    kctx,
	}, nil
}

type runner struct {
	connection api.Connection
	kubectl    kubectl.Kubectl
	kubectx    kubecfg.Kubectx
	client     kubernetes.Interface
}

func (r *runner) Create(spec Spec) (Pod, error) {
	commonMeta, err := createMetadata(r.connection)
	if err != nil {
		return nil, err
	}

	envSecretSpec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "thelma-sql-env-",
		},
		StringData: spec.Env,
	}
	commonMeta.write(&envSecretSpec.ObjectMeta)

	envSecret, err := r.client.CoreV1().Secrets(thelmaWorkloadsNamespace).Create(
		context.Background(),
		envSecretSpec,
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating env secret: %v", err)
	}
	log.Debug().Msgf("created env secret: %q", envSecret.Name)

	scriptSecretSpec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "thelma-sql-scripts-",
		},
		Data: spec.Scripts,
	}
	commonMeta.write(&scriptSecretSpec.ObjectMeta)
	scriptSecret, err := r.client.CoreV1().
		Secrets(thelmaWorkloadsNamespace).Create(
		context.Background(),
		scriptSecretSpec,
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating scripts secret: %v", err)
	}
	log.Debug().Msgf("created scripts secret: %q", scriptSecret.Name)

	podSpec := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "thelma-pod-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  clientContainerName,
					Image: "postgres:15",
					Command: []string{
						"/bin/sleep",
						strconv.Itoa(maxPodLifetimeSeconds),
					},
					EnvFrom: []v1.EnvFromSource{
						{
							SecretRef: &v1.SecretEnvSource{
								LocalObjectReference: v1.LocalObjectReference{
									Name: envSecret.Name,
								},
							},
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "scripts",
							ReadOnly:  true,
							MountPath: "/scripts",
						},
					},
					// TODO - limits/requests
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "scripts",
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName:  scriptSecret.Name,
							DefaultMode: utils.Nullable[int32](0755),
						},
					},
				},
			},
			ServiceAccountName: spec.ServiceAccount,
		},
	}
	if spec.Sidecar != nil {
		podSpec.Spec.Containers = append(podSpec.Spec.Containers, *spec.Sidecar)
	}
	commonMeta.write(&podSpec.ObjectMeta)

	_pod, err := r.client.CoreV1().Pods(thelmaWorkloadsNamespace).Create(
		context.Background(),
		podSpec,
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating pod: %v", err)
	}
	log.Debug().Msgf("created pod: %q", _pod.Name)

	log.Info().Msgf("Launched pod %s, waiting for it to become ready...", _pod.Name)
	if err = r.waitForPodToBeReady(_pod, podReadinessTimeout); err != nil {
		return nil, fmt.Errorf("error waiting for pod %s to be ready: %v", _pod.Name, err)
	}

	return &pod{
		name:          _pod.Name,
		envSecret:     envSecret.Name,
		scriptsSecret: scriptSecret.Name,
		kubectl:       r.kubectl,
		kubectx:       r.kubectx,
		client:        r.client,
	}, nil
}

func (r *runner) waitForPodToBeReady(_pod *v1.Pod, timeout time.Duration) error {
	log.Debug().Msgf("waiting up to %s for pod %s to become ready", timeout, _pod.Name)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	watch, err := r.client.CoreV1().Pods(thelmaWorkloadsNamespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", _pod.Name),
	})
	if err != nil {
		return fmt.Errorf("error creating watch for pod %s: %v", _pod.Name, err)
	}

	for event := range watch.ResultChan() {
		switch event.Object.(type) {
		case *v1.Pod:
		case error:
			log.Debug().Err(event.Object.(error)).Msgf("watch error")
			continue
		default:
			log.Debug().Msgf("could not decode event object")
			continue
		}

		p, ok := event.Object.(*v1.Pod)
		if !ok {
			log.Debug().Msgf("unexpected event object type: %#v", event.Object)
			continue
		}

		allContainersReady := true
		for _, c := range p.Status.ContainerStatuses {
			log.Debug().Msgf("container %s state: %s", c.Name, c.State.String())
			if !c.Ready {
				allContainersReady = false
			}
		}

		if allContainersReady {
			watch.Stop()
			log.Debug().Msgf("pod %s became ready", _pod.Name)
			return nil
		}
	}

	return fmt.Errorf("timed out after %s waiting for pod %s to be ready", timeout, _pod.Name)
}

func (r *runner) Cleanup() error {
	//TODO implement me
	panic("implement me")
}

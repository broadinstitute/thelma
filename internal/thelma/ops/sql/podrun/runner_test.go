package podrun

import (
	kubecfgmocks "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg/mocks"
	kubeclientsmocks "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/mocks"
	kubetesting "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/testing"
	kubemocks "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/testing/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	kubectlmocks "github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/maps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"strconv"
	"testing"
)

const project = "my-project"
const instance = "my-instance"
const serviceaccount = "my-sa"

type RunnerSuite struct {
	suite.Suite
	cluster   *statemocks.Cluster
	kubecfg   *kubecfgmocks.Kubeconfig
	kubectx   *kubecfgmocks.Kubectx
	kubectl   *kubectlmocks.Kubectl
	kubemocks *kubetesting.KubeMocks
	clients   *kubeclientsmocks.Clients
}

func (suite *RunnerSuite) SetupTest() {
	suite.cluster = &statemocks.Cluster{}

	suite.kubectx = &kubecfgmocks.Kubectx{}
	suite.kubectx.EXPECT().Namespace().Return(thelmaWorkloadsNamespace)

	suite.kubecfg = &kubecfgmocks.Kubeconfig{}
	suite.kubecfg.EXPECT().ForCluster(suite.cluster).Return(suite.kubectx, nil)

	suite.kubectl = kubectlmocks.NewKubectl(suite.T())

	suite.kubemocks = kubetesting.NewKubeMocks(thelmaWorkloadsNamespace)

	suite.clients = &kubeclientsmocks.Clients{}
	suite.clients.EXPECT().Kubecfg().Return(suite.kubecfg, nil)
	suite.clients.EXPECT().Kubectl().Return(suite.kubectl, nil)
	suite.clients.EXPECT().ForKubectx(suite.kubectx).Return(suite.kubemocks.KubeClient, nil)
}

func (suite *RunnerSuite) Test_Create_Exec_And_Delete() {
	suite.kubemocks.Secrets.EXPECT().
		Create(
			mock.Anything,
			mock.MatchedBy(func(s *corev1.Secret) bool {
				if s.GenerateName != "thelma-sql-env-" {
					return false
				}
				assert.Equal(suite.T(), map[string]string{"FOO": "BAR"}, s.StringData)
				return true
			}),
			metav1.CreateOptions{},
		).
		Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "thelma-sql-env-1"}}, nil)

	suite.kubemocks.Secrets.EXPECT().
		Create(
			mock.Anything,
			mock.MatchedBy(func(s *corev1.Secret) bool {
				if s.GenerateName != "thelma-sql-scripts-" {
					return false
				}
				data := maps.TransformValues(s.Data, func(v []byte) string {
					return string(v)
				})
				assert.Equal(suite.T(), map[string]string{"script.sh": "echo hello world"}, data)
				return true
			}),
			metav1.CreateOptions{},
		).
		Return(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "thelma-sql-scripts-1"}}, nil)

	suite.kubemocks.Pods.EXPECT().
		Create(
			mock.Anything,
			mock.MatchedBy(func(p *corev1.Pod) bool {
				if p.GenerateName != "thelma-pod-" {
					return false
				}
				assert.Equal(suite.T(), corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "scripts",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "thelma-sql-scripts-1",
									DefaultMode: utils.Nullable[int32](0755),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "sqlclient",
							Image: "my-image",
							Command: []string{
								"/bin/sleep",
								strconv.Itoa(maxPodLifetimeSeconds),
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									Prefix:       "",
									ConfigMapRef: nil,
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "thelma-sql-env-1",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "scripts",
									MountPath: "/my-scripts",
									ReadOnly:  true,
								},
							},
						},
						{
							Name:  "sidecar",
							Image: "sidecar-image",
						},
					},
					ServiceAccountName: serviceaccount,
				}, p.Spec)
				return true
			}),
			metav1.CreateOptions{},
		).
		Return(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "thelma-pod-1"},
		}, nil)

	readyEvent := watch.Event{
		Object: &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "sqlclient",
						Ready: true,
					},
					{
						Name:  "sidecar",
						Ready: true,
					},
				},
			},
		},
	}
	eventCh := make(chan watch.Event, 1)
	eventCh <- readyEvent
	close(eventCh)
	mockWatch := kubemocks.NewWatch(suite.T())
	mockWatch.EXPECT().Stop().Return()
	mockWatch.EXPECT().ResultChan().Return(eventCh)
	suite.kubemocks.Pods.EXPECT().Watch(mock.Anything, metav1.ListOptions{
		FieldSelector: "metadata.name=thelma-pod-1",
	}).Return(mockWatch, nil)

	container := kubectl.Container{Pod: "thelma-pod-1", Namespace: thelmaWorkloadsNamespace, Name: "sqlclient"}
	suite.kubectl.EXPECT().Exec(suite.kubectx, container, []string{"echo", "not interactive"}, mock.Anything).Return(nil)
	suite.kubectl.EXPECT().ExecInteractive(suite.kubectx, container, []string{"echo", "interactive"}).Return(nil)

	suite.kubemocks.Pods.EXPECT().Delete(mock.Anything, "thelma-pod-1", metav1.DeleteOptions{}).Return(nil)
	suite.kubemocks.Secrets.EXPECT().Delete(mock.Anything, "thelma-sql-scripts-1", metav1.DeleteOptions{}).Return(nil)
	suite.kubemocks.Secrets.EXPECT().Delete(mock.Anything, "thelma-sql-env-1", metav1.DeleteOptions{}).Return(nil)

	runner, err := New(api.Connection{
		Provider: api.Google,
		GoogleInstance: api.GoogleInstance{
			Project:      project,
			InstanceName: instance,
		},
		Options: api.ConnectionOptions{
			PrivilegeLevel: api.ReadWrite,
			ProxyCluster:   suite.cluster,
		},
	}, suite.clients)
	require.NoError(suite.T(), err)

	pod, err := runner.Create(Spec{
		DBMSSpec: DBMSSpec{
			ContainerImage: "my-image",
			Env:            map[string]string{"FOO": "BAR"},
			Scripts: map[string][]byte{
				"script.sh": []byte("echo hello world"),
			},
			ScriptsMount: "/my-scripts",
		},
		ProviderSpec: ProviderSpec{
			Sidecar: &corev1.Container{
				Name:  "sidecar",
				Image: "sidecar-image",
			},
			ServiceAccount: serviceaccount,
		},
	})
	require.NoError(suite.T(), err)

	require.NoError(suite.T(), pod.Exec([]string{"echo", "not interactive"}))
	require.NoError(suite.T(), pod.ExecInteractive([]string{"echo", "interactive"}))

	require.NoError(suite.T(), pod.Delete())
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(RunnerSuite))
}

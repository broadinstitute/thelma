package kubernetes

import (
	kubecfgmocks "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg/mocks"
	kubeclientsmocks "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/mocks"
	kubetesting "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/testing"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	kubectlmocks "github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

const cluster = "my-cluster"
const project = "my-project"
const database = "my-database"
const environment = "my-env"
const namespace = "terra-" + environment
const chart = "my-service"
const credsSecretName = "my-service-thelma-sql-secret"
const service = "my-postgres-service"

const fakeRoPass = "ropass"
const fakeRwPass = "rwpass"
const fakeGeneratedPass = "generated"

type KubernetesSuite struct {
	suite.Suite
	cluster   *statemocks.Cluster
	release   *statemocks.Release
	clients   *kubeclientsmocks.Clients
	kubecfg   *kubecfgmocks.Kubeconfig
	kubectx   *kubecfgmocks.Kubectx
	kubectl   *kubectlmocks.Kubectl
	kubemocks *kubetesting.KubeMocks
}

func (suite *KubernetesSuite) SetupTest() {
	suite.cluster = &statemocks.Cluster{}
	suite.cluster.EXPECT().Name().Return(cluster)
	suite.cluster.EXPECT().Project().Return(project)

	suite.release = &statemocks.Release{}
	suite.release.EXPECT().Namespace().Return(namespace)
	suite.release.EXPECT().Name().Return(chart)
	env := &statemocks.Environment{}
	env.EXPECT().Name().Return(environment)
	suite.release.EXPECT().Destination().Return(env)

	suite.kubecfg = &kubecfgmocks.Kubeconfig{}
	suite.kubectx = &kubecfgmocks.Kubectx{}
	suite.kubectx.EXPECT().Namespace().Return(namespace)

	suite.kubectl = kubectlmocks.NewKubectl(suite.T())

	suite.kubemocks = kubetesting.NewKubeMocks(namespace)

	suite.clients = &kubeclientsmocks.Clients{}
	suite.clients.EXPECT().Kubectl().Return(suite.kubectl, nil)
	suite.clients.EXPECT().Kubecfg().Return(suite.kubecfg, nil)
	suite.kubecfg.EXPECT().ForRelease(suite.release).Return(suite.kubectx, nil)
	suite.clients.EXPECT().ForKubectx(suite.kubectx).Return(suite.kubemocks.KubeClient, nil)
}

func (suite *KubernetesSuite) Test_Initialized_ReturnsFalse_IfNoSecret() {
	suite.expectListCredentialsSecret(false)

	provider := suite.buildProvider()
	initialized, err := provider.Initialized()
	require.NoError(suite.T(), err)
	assert.False(suite.T(), initialized)
}

func (suite *KubernetesSuite) Test_Initialized_ReturnsTrue_IfSecretExists() {
	suite.expectListCredentialsSecret(true)

	provider := suite.buildProvider()
	initialized, err := provider.Initialized()
	require.NoError(suite.T(), err)
	assert.True(suite.T(), initialized)
}

func (suite *KubernetesSuite) Test_Initialize() {
	suite.expectFeatureDetection()
	suite.expectListCredentialsSecret(false)
	suite.expectCreateCredentialsSecret()

	provider := suite.buildProvider()
	err := provider.Initialize()
	require.NoError(suite.T(), err)
}

func (suite *KubernetesSuite) Test_ClientSettings_ReadOnly() {
	suite.expectFeatureDetection()
	suite.expectGetCredentialsSecret()
	provider := suite.buildProvider()
	settings, err := provider.ClientSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), dbms.ClientSettings{
		Host:     service + "." + namespace + ".svc.cluster.local",
		Username: readonlyUsername,
		Password: fakeRoPass,
		Database: database,
		Init: dbms.InitSettings{
			CreateUsers: true,
			ReadOnlyUser: dbms.InitUser{
				Name:     readonlyUsername,
				Password: fakeRoPass,
			},
			ReadWriteUser: dbms.InitUser{
				Name:     readwriteUsername,
				Password: fakeRwPass,
			},
		},
	}, settings)
}

func (suite *KubernetesSuite) Test_ClientSettings_ReadWrite() {
	suite.expectFeatureDetection()
	suite.expectGetCredentialsSecret()
	provider := suite.buildProvider()
	settings, err := provider.ClientSettings(func(options *api.ConnectionOptions) {
		options.PrivilegeLevel = api.ReadWrite
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), dbms.ClientSettings{
		Host:     service + "." + namespace + ".svc.cluster.local",
		Username: readwriteUsername,
		Password: fakeRwPass,
		Database: database,
		Init: dbms.InitSettings{
			CreateUsers: true,
			ReadOnlyUser: dbms.InitUser{
				Name:     readonlyUsername,
				Password: fakeRoPass,
			},
			ReadWriteUser: dbms.InitUser{
				Name:     readwriteUsername,
				Password: fakeRwPass,
			},
		},
	}, settings)
}

func (suite *KubernetesSuite) Test_ClientSettings_Admin() {
	suite.expectFeatureDetection()
	suite.expectGetCredentialsSecret()
	suite.expectResetAdminPassword()

	provider := suite.buildProvider()
	settings, err := provider.ClientSettings(func(options *api.ConnectionOptions) {
		options.PrivilegeLevel = api.Admin
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), dbms.ClientSettings{
		Host:     service + "." + namespace + ".svc.cluster.local",
		Username: api.Postgres.AdminUser(),
		Password: fakeGeneratedPass,
		Database: database,
		Init: dbms.InitSettings{
			CreateUsers: true,
			ReadOnlyUser: dbms.InitUser{
				Name:     readonlyUsername,
				Password: fakeRoPass,
			},
			ReadWriteUser: dbms.InitUser{
				Name:     readwriteUsername,
				Password: fakeRwPass,
			},
		},
	}, settings)
}

func (suite *KubernetesSuite) Test_PodSpec() {
	suite.expectFeatureDetection()
	provider := suite.buildProvider()
	spec, err := provider.PodSpec()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), serviceAccountName, spec.ServiceAccount)
}

func (suite *KubernetesSuite) Test_DetectDBMS_DetectsPostgres() {
	suite.expectFeatureDetection()
	provider := suite.buildProvider()
	dbms, err := provider.DetectDBMS()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), api.Postgres, dbms)
}

func TestKubernetesSuite(t *testing.T) {
	suite.Run(t, new(KubernetesSuite))
}

func (suite *KubernetesSuite) expectResetAdminPassword() {
	suite.kubectl.EXPECT().Exec(
		suite.kubectx,
		kubectl.Container{
			Pod:       "postgres-0",
			Name:      "postgres",
			Namespace: namespace,
		}, []string{
			"psql", "--no-psqlrc", "--host", "localhost", "-U", "postgres", "-c", "alter user postgres password '" + fakeGeneratedPass + "';",
		},
	).Return(nil)
}

func (suite *KubernetesSuite) expectCreateCredentialsSecret() {
	suite.kubemocks.Secrets.EXPECT().Create(
		mock.Anything,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: credsSecretName},
			StringData: map[string]string{
				readonlyUsername:  fakeGeneratedPass,
				readwriteUsername: fakeGeneratedPass,
			},
		},
		metav1.CreateOptions{},
	).Return(
		&corev1.Secret{}, nil,
	)
}

func (suite *KubernetesSuite) expectGetCredentialsSecret() {
	s := mockCredsSecret()

	suite.kubemocks.Secrets.EXPECT().
		Get(mock.Anything, credsSecretName, metav1.GetOptions{}).
		Return(&s, nil)
}

func (suite *KubernetesSuite) expectListCredentialsSecret(exists bool) {
	var items []corev1.Secret
	if exists {
		items = append(items, mockCredsSecret())
	}

	suite.kubemocks.Secrets.EXPECT().
		List(mock.Anything, metav1.ListOptions{
			FieldSelector: "metadata.name=" + credsSecretName,
		}).
		Return(&corev1.SecretList{
			Items: items,
		}, nil)
}

func mockCredsSecret() corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: credsSecretName,
		},
		Data: map[string][]byte{
			readonlyUsername:  []byte(fakeRoPass),
			readwriteUsername: []byte(fakeRwPass),
		},
	}
}

func (suite *KubernetesSuite) expectFeatureDetection() {
	suite.kubemocks.StatefulSets.EXPECT().
		List(mock.Anything, metav1.ListOptions{
			LabelSelector: "argocd.argoproj.io/instance=my-service-my-env",
		}).
		Return(&appsv1.StatefulSetList{
			Items: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-service-postgres-sts",
					},
					Spec: appsv1.StatefulSetSpec{Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"sts": "postgres",
						},
					}},
				},
			},
		}, nil)

	suite.kubemocks.Pods.EXPECT().
		List(mock.Anything, metav1.ListOptions{
			LabelSelector: "sts=postgres",
		}).
		Return(&corev1.PodList{
			Items: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "postgres-0",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "postgres",
							},
						},
					},
				},
			},
		}, nil)

	suite.kubemocks.Services.EXPECT().
		List(mock.Anything, metav1.ListOptions{
			LabelSelector: "argocd.argoproj.io/instance=my-service-my-env",
		}).
		Return(&corev1.ServiceList{Items: []corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: service,
				},
			},
		}}, nil)
}

func (suite *KubernetesSuite) buildProvider() provider.Provider {
	conn := api.Connection{
		Provider: api.Kubernetes,
		KubernetesInstance: api.KubernetesInstance{
			Release: suite.release,
		},
		Options: api.ConnectionOptions{
			Database:       database,
			PrivilegeLevel: api.ReadOnly,
			ProxyCluster:   suite.cluster,
			Release:        suite.release,
			Shell:          false,
		},
	}
	p, err := newKubernetesProvider(conn, suite.clients, &fixedPwg{})
	require.NoError(suite.T(), err)
	return p
}

type fixedPwg struct {
}

func (f *fixedPwg) Generate() string {
	return fakeGeneratedPass
}

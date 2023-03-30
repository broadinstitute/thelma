package google

import (
	sqladminmocks "github.com/broadinstitute/thelma/internal/thelma/clients/google/sqladmin/mocks"
	vaulttesting "github.com/broadinstitute/thelma/internal/thelma/clients/vault/testing"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/sqladmin/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

const project = "my-project"
const instance = "my-instance"
const region = "us-central1"
const database = "my-database"
const cluster = "my-cluster"
const roUser = "thelma-sql-ro-" + cluster + "@" + project + ".iam"
const rwUser = "thelma-sql-rw-" + cluster + "@" + project + ".iam"

const postgres15Version = "POSTGRES_15"

// init settings that will always be created for a postgres 15 instance
var postgres15InitSettings = dbms.InitSettings{
	CreateUsers: false,
	ReadOnlyUser: dbms.InitUser{
		Name:     roUser,
		Password: "",
	},
	ReadWriteUser: dbms.InitUser{
		Name:     rwUser,
		Password: "",
	},
}

type GoogleSuite struct {
	suite.Suite
	proxyCluster   *statemocks.Cluster
	sqladminClient *sqladminmocks.Client
	vaultServer    *vaulttesting.FakeVaultServer
	vaultClient    *vaultapi.Client
}

func (suite *GoogleSuite) SetupTest() {
	suite.proxyCluster = &statemocks.Cluster{}
	suite.proxyCluster.EXPECT().Name().Return("my-cluster")
	suite.proxyCluster.EXPECT().Project().Return(project)

	suite.sqladminClient = sqladminmocks.NewClient(suite.T())
	suite.vaultServer = vaulttesting.NewFakeVaultServer(suite.T())
	suite.vaultClient = suite.vaultServer.GetClient()
}

func (suite *GoogleSuite) Test_Initialized_ReturnsFalseIfUsersDoNotExist() {
	suite.sqladminClient.EXPECT().GetInstanceLocalUsers(project, instance).Return([]string{"should-be-ignored"}, nil)
	initialized, err := suite.buildProvider().Initialized()
	require.NoError(suite.T(), err)
	assert.False(suite.T(), initialized)
}

func (suite *GoogleSuite) Test_Initialized_ReturnsTrueIfUsersExist() {
	suite.sqladminClient.EXPECT().GetInstanceLocalUsers(project, instance).Return([]string{
		roUser,
		rwUser,
	}, nil)
	initialized, err := suite.buildProvider().Initialized()
	require.NoError(suite.T(), err)
	assert.True(suite.T(), initialized)
}

func (suite *GoogleSuite) Test_DetectDBMS_ReturnsErrorIfNotMatched() {
	suite.expectGetInstanceWith("SQLSERVER_2019_EXPRESS", false)
	_, err := suite.buildProvider().DetectDBMS()
	require.Error(suite.T(), err)
}

func (suite *GoogleSuite) Test_DetectDBMS_IdentifiesPostgres() {
	suite.expectGetInstanceWith(postgres15Version, true)

	dbms, err := suite.buildProvider().DetectDBMS()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), api.Postgres, dbms)
}

func (suite *GoogleSuite) Test_Initialize_EnablesIAMAndAddsUsers() {
	suite.expectGetInstanceWith(postgres15Version, false)
	suite.sqladminClient.EXPECT().PatchInstance(project, instance, &sqladmin.DatabaseInstance{
		Settings: &sqladmin.Settings{DatabaseFlags: []*sqladmin.DatabaseFlags{
			{
				Name:  cloudSqlIamAuthenticationFlag,
				Value: cloudSqlFlagEnabled,
			},
		}},
	}).Return(nil)
	suite.sqladminClient.EXPECT().GetInstanceLocalUsers(project, instance).Return([]string{}, nil)
	suite.sqladminClient.EXPECT().AddUser(project, instance, &sqladmin.User{Name: roUser, Type: cloudSqlAccountTypeIAM}).Return(nil)
	suite.sqladminClient.EXPECT().AddUser(project, instance, &sqladmin.User{Name: rwUser, Type: cloudSqlAccountTypeIAM}).Return(nil)

	require.NoError(suite.T(), suite.buildProvider().Initialize())
}

func (suite *GoogleSuite) Test_ClientSettings_ReadOnlyUser() {
	suite.expectGetInstanceWith(postgres15Version, true)

	settings, err := suite.buildProvider().ClientSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), dbms.ClientSettings{
		Username: roUser,
		Password: "",
		Host:     cloudsqlProxySidecarAddress,
		Database: database,
		Nickname: "thelma-sql-ro",
		Init:     postgres15InitSettings,
	}, settings)
}

func (suite *GoogleSuite) Test_ClientSettings_ReadWriteUser() {
	suite.expectGetInstanceWith(postgres15Version, true)

	provider := suite.buildProvider(func(options *api.ConnectionOptions) {
		options.PermissionLevel = api.ReadWrite
	})

	settings, err := provider.ClientSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), dbms.ClientSettings{
		Username: rwUser,
		Password: "",
		Host:     cloudsqlProxySidecarAddress,
		Database: database,
		Nickname: "thelma-sql-rw",
		Init:     postgres15InitSettings,
	}, settings)
}

func (suite *GoogleSuite) Test_ClientSettings_AdminUser() {
	suite.expectGetInstanceWith(postgres15Version, true)

	var generatedPassword string
	suite.sqladminClient.EXPECT().ResetPassword(project, instance, api.Postgres.AdminUser(), mock.Anything).Run(func(_ string, _ string, _ string, password string) {
		generatedPassword = password
	}).Return(nil)

	settings, err := suite.buildProvider(func(options *api.ConnectionOptions) {
		options.PermissionLevel = api.Admin
	}).ClientSettings()

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), dbms.ClientSettings{
		Username: api.Postgres.AdminUser(),
		Password: generatedPassword,
		Host:     cloudsqlProxySidecarAddress,
		Database: database,
		Nickname: "",
		Init:     postgres15InitSettings,
	}, settings)
}

func (suite *GoogleSuite) Test_PodSpec_RoUser() {
	suite.expectGetInstanceWith(postgres15Version, true)

	spec, err := suite.buildProvider(func(options *api.ConnectionOptions) {
		options.PermissionLevel = api.ReadOnly
	}).PodSpec()

	require.NoError(suite.T(), err)

	sidecar := defaultSidecar()
	sidecar.Command = append(sidecar.Command, "-enable_iam_login")

	assert.Equal(suite.T(), podrun.ProviderSpec{
		Sidecar:        sidecar,
		ServiceAccount: "thelma-sql-ro", // k8s service account mapped by WI to the thelma-sql-ro IAM account for the cluster
	}, spec)
}

func (suite *GoogleSuite) Test_PodSpec_RwUser() {
	suite.expectGetInstanceWith(postgres15Version, true)

	spec, err := suite.buildProvider(func(options *api.ConnectionOptions) {
		options.PermissionLevel = api.ReadWrite
	}).PodSpec()

	require.NoError(suite.T(), err)

	sidecar := defaultSidecar()
	sidecar.Command = append(sidecar.Command, "-enable_iam_login")

	assert.Equal(suite.T(), podrun.ProviderSpec{
		Sidecar:        sidecar,
		ServiceAccount: "thelma-sql-rw", // k8s service account mapped by WI to the thelma-sql-rw IAM account for the cluster
	}, spec)
}

func (suite *GoogleSuite) Test_PodSpec_AdminUser() {
	suite.expectGetInstanceWith(postgres15Version, true)

	spec, err := suite.buildProvider(func(options *api.ConnectionOptions) {
		options.PermissionLevel = api.Admin
	}).PodSpec()

	require.NoError(suite.T(), err)

	sidecar := defaultSidecar()
	// note we don't enable iam login in the cloud sql proxy arguments, since we
	// log in as the Postgres user

	assert.Equal(suite.T(), podrun.ProviderSpec{
		Sidecar: sidecar,
		// we _do_ however, run as a WI service account that has database connect permissions
		ServiceAccount: "thelma-sql-ro",
	}, spec)
}

func TestGoogleSuite(t *testing.T) {
	suite.Run(t, new(GoogleSuite))
}

func (suite *GoogleSuite) buildProvider(opts ...func(options *api.ConnectionOptions)) provider.Provider {
	options := api.ConnectionOptions{
		Database:        database,
		PermissionLevel: api.ReadOnly,
		ProxyCluster:    suite.proxyCluster,
		Release:         nil,
		Shell:           false,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return New(api.Connection{
		Provider: api.Google,
		GoogleInstance: api.GoogleInstance{
			Project:      project,
			InstanceName: instance,
		},
		Options: options,
	}, suite.sqladminClient, suite.vaultClient)
}

func (suite *GoogleSuite) expectGetInstanceWith(databaseVersion string, iamEnabled bool) {
	var flags []*sqladmin.DatabaseFlags
	if iamEnabled {
		flags = append(flags, &sqladmin.DatabaseFlags{
			Name:  cloudSqlIamAuthenticationFlag,
			Value: cloudSqlFlagEnabled,
		})
	}

	suite.sqladminClient.EXPECT().GetInstance(project, instance).Return(&sqladmin.DatabaseInstance{
		Name:            instance,
		DatabaseVersion: databaseVersion,
		Settings: &sqladmin.Settings{
			DatabaseFlags: flags,
		},
		Region: region,
	}, nil)
}

func defaultSidecar() *v1.Container {
	return &v1.Container{
		Name:  "sqlproxy",
		Image: "gcr.io/cloudsql-docker/gce-proxy:latest",
		Command: []string{
			"/cloud_sql_proxy",
			"-instances=$(SQL_INSTANCE_PROJECT):$(SQL_INSTANCE_REGION):$(SQL_INSTANCE_NAME)=tcp:5432",
			"-use_http_health_check",
			"-health_check_port=9090",
			"-verbose",
		},
		Env: []v1.EnvVar{
			{
				Name:  "SQL_INSTANCE_PROJECT",
				Value: project,
			},
			{
				Name:  "SQL_INSTANCE_REGION",
				Value: region,
			}, {
				Name:  "SQL_INSTANCE_NAME",
				Value: instance,
			},
		},
		LivenessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/liveness",
					Port: intstr.FromInt(9090),
				},
			},
			PeriodSeconds:    60,
			TimeoutSeconds:   30,
			FailureThreshold: 5,
		},
		ReadinessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/readiness",
					Port: intstr.FromInt(9090),
				},
			},
			PeriodSeconds:    10,
			TimeoutSeconds:   5,
			SuccessThreshold: 1,
			FailureThreshold: 1,
		},
		StartupProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/startup",
					Port: intstr.FromInt(9090),
				},
			},
			InitialDelaySeconds: 0,
			TimeoutSeconds:      5,
			PeriodSeconds:       1,
			SuccessThreshold:    0,
			FailureThreshold:    20,
		},
	}
}

package connector

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	dbmsmocks "github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	podrunmocks "github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun/mocks"
	providermocks "github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ConnectorSuite struct {
	suite.Suite
	provider *providermocks.Provider
	dbms     *dbmsmocks.DBMS
	runner   *podrunmocks.Runner
}

func (suite *ConnectorSuite) SetupTest() {
	suite.provider = providermocks.NewProvider(suite.T())
	suite.dbms = dbmsmocks.NewDBMS(suite.T())
	suite.runner = podrunmocks.NewRunner(suite.T())
}

func (suite *ConnectorSuite) Test_Connect() {
	suite.provider.EXPECT().Initialized().Return(true, nil)
	settings := dbms.ClientSettings{}
	suite.provider.EXPECT().ClientSettings().Return(settings, nil)
	dspec := podrun.DBMSSpec{}
	pspec := podrun.ProviderSpec{}
	suite.dbms.EXPECT().PodSpec(settings).Return(dspec, nil)
	suite.provider.EXPECT().PodSpec().Return(pspec, nil)
	pod := podrunmocks.NewPod(suite.T())
	suite.runner.EXPECT().Create(podrun.Spec{DBMSSpec: dspec, ProviderSpec: pspec}).Return(pod, nil)
	suite.dbms.EXPECT().ShellCommand().Return([]string{"dbms-shell"})
	pod.EXPECT().ExecInteractive([]string{"dbms-shell"}).Return(nil)
	pod.EXPECT().Close().Return(nil)
	conn := newConnector(api.Connection{
		Provider: api.Google,
		GoogleInstance: api.GoogleInstance{
			Project:      "my-project",
			InstanceName: "my-instance",
		},
		Options: api.ConnectionOptions{
			Database:       "my-db",
			PrivilegeLevel: api.ReadOnly,
			ProxyCluster:   nil,
			Shell:          false,
		},
	}, suite.provider, suite.runner, suite.dbms)
	require.NoError(suite.T(), conn.Connect())
}

func TestConnector(t *testing.T) {
	suite.Run(t, new(ConnectorSuite))
}

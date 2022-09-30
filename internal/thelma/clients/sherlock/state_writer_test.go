package sherlock_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type sherlockStateWriterClientSuite struct {
	suite.Suite
	state          terra.State
	conflictServer *httptest.Server
}

func Test_SherlockStateWriterClient(t *testing.T) {
	suite.Run(t, new(sherlockStateWriterClientSuite))
}

func (suite *sherlockStateWriterClientSuite) SetupSuite() {
	suite.state = constructFakeState(suite.T())
	suite.conflictServer = newMockConflictServer()
}

func (suite *sherlockStateWriterClientSuite) TearDownSuite() {
	suite.conflictServer.Close()
}

func (suite *sherlockStateWriterClientSuite) TestIgnore409Conflict() {
	client, err := sherlock.NewWithHostnameOverride(suite.conflictServer.URL, "")
	suite.Assert().NoError(err)

	stateClusters, err := suite.state.Clusters().All()
	suite.Assert().NoError(err)
	err = client.WriteClusters(stateClusters)
	suite.Assert().NoError(err)

	stateEnvironments, err := suite.state.Environments().All()
	suite.Assert().NoError(err)
	err = client.WriteEnvironments(stateEnvironments)
	suite.Assert().NoError(err)
}

func constructFakeState(t *testing.T) terra.State {
	builder := builder.NewBuilder().WithTestDefaults(t)
	app, err := builder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)
	return state
}

func newMockConflictServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/charts", mock409ConflictHandler())
	mux.HandleFunc("/api/v2/environments", mock409ConflictHandler())
	mux.HandleFunc("/api/v2/clusters", mock409ConflictHandler())
	mux.HandleFunc("/api/v2/chart-releases", mock409ConflictHandler())
	return httptest.NewServer(mux)
}

func mock409ConflictHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}
}

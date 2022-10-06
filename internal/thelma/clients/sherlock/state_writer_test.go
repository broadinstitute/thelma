package sherlock_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/broadinstitute/sherlock/clients/go/client/environments"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type sherlockStateWriterClientSuite struct {
	suite.Suite
	state                  terra.State
	conflictServer         *httptest.Server
	errServer              *httptest.Server
	successfulCreateServer *httptest.Server
}

func Test_SherlockStateWriterClient(t *testing.T) {
	suite.Run(t, new(sherlockStateWriterClientSuite))
}

func (suite *sherlockStateWriterClientSuite) SetupSuite() {
	suite.state = constructFakeState(suite.T())
	suite.conflictServer = newMockConflictServer()
	suite.errServer = newMockErroringSherlockServer()
	suite.successfulCreateServer = newMockSuccessfulCreateServer()
}

func (suite *sherlockStateWriterClientSuite) TearDownSuite() {
	suite.conflictServer.Close()
	suite.errServer.Close()
	suite.successfulCreateServer.Close()
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
	_, err = client.WriteEnvironments(stateEnvironments)
	suite.Assert().NoError(err)
}

func (suite *sherlockStateWriterClientSuite) TestPropagatesServerError() {
	client, err := sherlock.NewWithHostnameOverride(suite.errServer.URL, "")
	suite.Assert().NoError(err)

	stateClusters, err := suite.state.Clusters().All()
	suite.Assert().NoError(err)
	err = client.WriteClusters(stateClusters)
	suite.Assert().Error(err)

	stateEnvironments, err := suite.state.Environments().All()
	suite.Assert().NoError(err)
	_, err = client.WriteEnvironments(stateEnvironments)
	suite.Assert().Error(err)
}

func (suite *sherlockStateWriterClientSuite) TestSuccessfulStateExport() {
	client, err := sherlock.NewWithHostnameOverride(suite.successfulCreateServer.URL, "")
	suite.Assert().NoError(err)

	stateClusters, err := suite.state.Clusters().All()
	suite.Assert().NoError(err)
	err = client.WriteClusters(stateClusters)
	suite.Assert().NoError(err)

	stateEnvironments, err := suite.state.Environments().All()
	suite.Assert().NoError(err)
	_, err = client.WriteEnvironments(stateEnvironments)
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

func newMockErroringSherlockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments", mockErroringHandler())
	mux.HandleFunc("/api/v2/clusters", mockErroringHandler())
	mux.HandleFunc("/api/v2/chart-releases", mockErroringHandler())
	mux.HandleFunc("/api/v2/charts", mockErroringHandler())
	return httptest.NewServer(mux)
}

func newMockSuccessfulCreateServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments", mockSuccessfulCreateHandler())
	mux.HandleFunc("/api/v2/clusters", mockSuccessfulCreateHandler())
	mux.HandleFunc("/api/v2/chart-releases", mockSuccessfulCreateHandler())
	mux.HandleFunc("/api/v2/charts", mockSuccessfulCreateHandler())
	return httptest.NewServer(mux)
}

func mock409ConflictHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}
}

func mockErroringHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func mockSuccessfulCreateHandler() http.HandlerFunc {
	response := environments.NewPostAPIV2EnvironmentsCreated()
	payload := &models.V2controllersEnvironment{
		Name: "test-env",
	}
	response.Payload = payload

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(&response)
	}
}

package sherlock_test

import (
	"encoding/json"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/broadinstitute/sherlock/sherlock-go-client/client/environments"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type sherlockStateWriterClientSuite struct {
	suite.Suite
	state                  terra.State
	conflictServer         *httptest.Server
	errServer              *httptest.Server
	successfulCreateServer *httptest.Server
	successfulDeleteServer *httptest.Server
	errDeleteServer        *httptest.Server
}

func Test_SherlockStateWriterClient(t *testing.T) {
	suite.Run(t, new(sherlockStateWriterClientSuite))
}

func (suite *sherlockStateWriterClientSuite) SetupSuite() {
	suite.state = constructFakeState(suite.T())
	suite.conflictServer = newMockConflictServer()
	suite.errServer = newMockErroringSherlockServer()
	suite.successfulCreateServer = newMockSuccessfulCreateServer()
	suite.successfulDeleteServer = newMockSuccessfulDeleteServer()
	suite.errDeleteServer = newMockErroringDeleteServer()
}

func (suite *sherlockStateWriterClientSuite) TearDownSuite() {
	suite.conflictServer.Close()
	suite.errServer.Close()
	suite.successfulCreateServer.Close()
	suite.successfulDeleteServer.Close()
	suite.errDeleteServer.Close()
}

func (suite *sherlockStateWriterClientSuite) TestIgnore409Conflict() {
	client, err := sherlock.NewClient("", func(options *sherlock.Options) {
		options.Addr = suite.conflictServer.URL
	})
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
	client, err := sherlock.NewClient("", func(options *sherlock.Options) {
		options.Addr = suite.errServer.URL
	})
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
	client, err := sherlock.NewClient("", func(options *sherlock.Options) {
		options.Addr = suite.successfulCreateServer.URL
	})
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

func (suite *sherlockStateWriterClientSuite) TestSuccessfulDelete() {
	mockEnv := mocks.NewEnvironment(suite.T())
	mockEnv.On("Name").Return("deleted-env")
	mockEnv.On("Releases").Return(nil)
	client, err := sherlock.NewClient("", func(options *sherlock.Options) {
		options.Addr = suite.successfulDeleteServer.URL
	})

	suite.Assert().NoError(err)
	_, err = client.DeleteEnvironments([]terra.Environment{mockEnv})
	suite.Assert().NoError(err)
}

func (suite *sherlockStateWriterClientSuite) TestErrorOnDelete() {
	mockEnv := mocks.NewEnvironment(suite.T())
	mockEnv.On("Name").Return("deleted-env")
	mockEnv.On("Releases").Return(nil)
	client, err := sherlock.NewClient("", func(options *sherlock.Options) {
		options.Addr = suite.errDeleteServer.URL
	})

	suite.Assert().NoError(err)
	_, err = client.DeleteEnvironments([]terra.Environment{mockEnv})
	suite.Assert().Error(err)
}

func constructFakeState(t *testing.T) terra.State {
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)
	builder := builder.NewBuilder().WithTestDefaults(t).UseCustomStateLoader(fixture.Mocks().StateLoader)
	app, err := builder.Build()
	require.NoError(t, err)

	state, err := app.State()
	require.NoError(t, err)
	return state
}

func newMockConflictServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/charts/v3", mock409ConflictHandler())
	mux.HandleFunc("/api/v2/environments", mock409ConflictHandler())
	mux.HandleFunc("/api/clusters/v3", mock409ConflictHandler())
	mux.HandleFunc("/api/v2/chart-releases", mock409ConflictHandler())
	return httptest.NewServer(mux)
}

func newMockErroringSherlockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments", mockErroringHandler())
	mux.HandleFunc("/api/clusters/v3", mockErroringHandler())
	mux.HandleFunc("/api/v2/chart-releases", mockErroringHandler())
	mux.HandleFunc("/api/charts/v3", mockErroringHandler())
	return httptest.NewServer(mux)
}

func newMockSuccessfulCreateServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments", mockSuccessfulCreateHandler())
	mux.HandleFunc("/api/clusters/v3", mockSuccessfulCreateHandler())
	mux.HandleFunc("/api/v2/chart-releases", mockSuccessfulCreateHandler())
	mux.HandleFunc("/api/charts/v3", mockSuccessfulCreateHandler())
	return httptest.NewServer(mux)
}

func newMockSuccessfulDeleteServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments/deleted-env", mockDeleteEnvironmentsHandler())
	return httptest.NewServer(mux)
}

func newMockErroringDeleteServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments/deleted-env", mockErroringHandler())
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
		_ = json.NewEncoder(w).Encode(&response)
	}
}

func mockDeleteEnvironmentsHandler() http.HandlerFunc {
	response := environments.NewDeleteAPIV2EnvironmentsSelectorOK()
	payload := &models.V2controllersEnvironment{
		Name: "deleted-env",
	}
	response.Payload = payload

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&response)
	}

}

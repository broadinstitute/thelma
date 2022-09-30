package sherlock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/broadinstitute/sherlock/clients/go/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type mockOkResponse struct {
	Ok bool
}

// Verify that the sherlock client can successfully issue a request against a mock sherlock backend
func Test_NewClient(t *testing.T) {
	mockSherlockServer := httptest.NewServer(newMockSherlockStatusHandler())
	defer mockSherlockServer.Close()

	thelmaConfig, err := config.Load(config.WithTestDefaults(t), config.WithOverride("sherlock.addr", mockSherlockServer.URL))
	require.NoError(t, err)

	client, err := New(thelmaConfig, "fake")
	require.NoError(t, err)

	err = client.getStatus()
	require.NoError(t, err)
}

func newMockSherlockStatusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOkResponse{Ok: true})
	})
}

func TestSherlockClientSuite(t *testing.T) {
	suite.Run(t, new(sherlockClientSuite))
}

type sherlockClientSuite struct {
	suite.Suite
	server    *httptest.Server
	errServer *httptest.Server
	config    config.Config
	errConfig config.Config
}

func (suite *sherlockClientSuite) SetupSuite() {
	suite.server = newMockSherlockServer()
	suite.errServer = newMockErroringSherlockServer()
	serverConfig, err := config.Load(
		config.WithTestDefaults(suite.T()),
		config.WithOverride("sherlock.addr", suite.server.URL),
	)
	suite.Assert().NoError(err)

	errConfig, err := config.Load(
		config.WithTestDefaults(suite.T()),
		config.WithOverride("sherlock.addr", suite.errServer.URL),
	)
	suite.Assert().NoError(err)
	suite.config = serverConfig
	suite.errConfig = errConfig
}

func (suite *sherlockClientSuite) TearDownSuite() {
	suite.server.Close()
	suite.errServer.Close()
}

func (suite *sherlockClientSuite) TestFetchEnvironments() {
	client, err := New(suite.config, "fake")
	suite.Assert().NoError(err)
	envs, err := client.Environments()
	suite.Assert().NoError(err)

	suite.Assert().Equal(2, len(envs))
	suite.Assert().Equal("dev", envs[0].Name)
}

func (suite *sherlockClientSuite) TestFetchClusters() {
	client, err := New(suite.config, "fake")
	suite.Assert().NoError(err)
	clusters, err := client.Clusters()
	suite.Assert().NoError(err)

	suite.Assert().Equal(2, len(clusters))
	suite.Assert().Equal("5.6.7.8", clusters[1].Address)
}

func (suite *sherlockClientSuite) TestFetchEnvironmentReleases() {
	client, err := New(suite.config, "fake")
	suite.Assert().NoError(err)
	envReleases, err := client.EnvironmentReleases("dev")
	suite.Assert().NoError(err)

	suite.Assert().Equal(1, len(envReleases))
	suite.Assert().Equal("sam", envReleases[0].Chart)
}

func (suite *sherlockClientSuite) TestFetchClusterReleases() {
	client, err := New(suite.config, "fake")
	suite.Assert().NoError(err)
	clusterReleases, err := client.ClusterReleases("tools")
	suite.Assert().NoError(err)

	suite.Assert().Equal(2, len(clusterReleases))
	suite.Assert().Equal("argocd-tools", clusterReleases[1].Name)
}

func (suite *sherlockClientSuite) TestFetchEnvironmentsError() {
	client, err := New(suite.errConfig, "fake")
	suite.Assert().NoError(err)
	_, err = client.Environments()
	suite.Assert().Error(err)
}

func (suite *sherlockClientSuite) TestFetchClustersError() {
	client, err := New(suite.errConfig, "fake")
	suite.Assert().NoError(err)
	_, err = client.Clusters()
	suite.Assert().Error(err)
}

func (suite *sherlockClientSuite) TestFetchEnvironmentReleasesError() {
	client, err := New(suite.errConfig, "fake")
	suite.Assert().NoError(err)
	_, err = client.EnvironmentReleases("blah")
	suite.Assert().Error(err)
}

func (suite *sherlockClientSuite) TestFetchClusterReleasesError() {
	client, err := New(suite.errConfig, "fake")
	suite.Assert().NoError(err)
	_, err = client.ClusterReleases("blah")
	suite.Assert().Error(err)
}

func newMockSherlockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments", mockEnvironmentsHandler())
	mux.HandleFunc("/api/v2/clusters", mockClustersHandler())
	mux.HandleFunc("/api/v2/chart-releases", mockChartReleasesHandler())
	return httptest.NewServer(mux)
}

func newMockErroringSherlockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments", mockErroringHandler())
	mux.HandleFunc("/api/v2/clusters", mockErroringHandler())
	mux.HandleFunc("/api/v2/chart-releases", mockErroringHandler())
	return httptest.NewServer(mux)
}

func mockEnvironmentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]*models.V2controllersEnvironment{
			{
				Base:           "live",
				DefaultCluster: "terra-dev",
				Name:           "dev",
			},
			{
				Base:           "live",
				DefaultCluster: "terra-prod",
				Name:           "prod",
			},
		})
	}
}

func mockClustersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]*models.V2controllersCluster{
			{
				Address:       "1.2.3.4",
				Base:          "terra",
				GoogleProject: "dev-proj",
				Name:          "terra-dev",
			},
			{
				Address:       "5.6.7.8",
				Base:          "tools",
				GoogleProject: "tools-proj",
				Name:          "tools",
			},
		})
	}
}

func mockChartReleasesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if r.URL.Query().Has("environment") {
			environmentParam := r.URL.Query().Get("environment")
			if environmentParam == "dev" {
				_ = encoder.Encode([]*models.V2controllersChartRelease{
					{
						AppVersionExact:   "0.2.1",
						Chart:             "sam",
						ChartVersionExact: "1.2.3",
						Name:              "sam-dev",
						Cluster:           "terra-dev",
						Environment:       "dev",
					},
				})
			}
		}
		if r.URL.Query().Has("cluster") {
			clusterParam := r.URL.Query().Get("cluster")
			if clusterParam == "tools" {
				_ = encoder.Encode([]*models.V2controllersChartRelease{
					{
						AppVersionExact:   "2.2.1",
						Chart:             "grafana",
						ChartVersionExact: "0.0.3",
						Name:              "grafana-tools",
						Cluster:           "tools",
					},
					{
						AppVersionExact:   "1.0.5",
						Chart:             "argocd",
						ChartVersionExact: "0.3.1",
						Name:              "argocd-tools",
						Cluster:           "tools",
					},
				})
			}
		}
	}
}

func mockErroringHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

package sherlock

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type mockOkResponse struct {
	Ok bool
}

// Verify that the sherlock client can successfully issue a request against a mock sherlock backend
func Test_NewClient(t *testing.T) {
	testGhaToken := "test github actions oidc jwt"
	testIapToken := "test identity aware proxy token"

	t.Run("just iap", func(t *testing.T) {
		mockSherlockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, fmt.Sprintf("Bearer %s", testIapToken), r.Header.Get("Authorization"))
			require.Empty(t, r.Header.Get(sherlockGhaOidcHeader))
			w.Header().Add("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockOkResponse{Ok: true})
		}))
		defer mockSherlockServer.Close()

		thelmaConfig, err := config.Load(config.WithTestDefaults(t), config.WithOverride("sherlock.addr", mockSherlockServer.URL))
		require.NoError(t, err)

		client, err := NewClient(testIapToken, func(options *Options) {
			options.ConfigSource = thelmaConfig
			options.GhaOidcTokenProvider = &credentials.MockTokenProvider{ReturnNil: true}
		})
		require.NoError(t, err)

		err = client.GetStatus()
		require.NoError(t, err)
	})

	t.Run("iap and gha", func(t *testing.T) {
		mockSherlockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, fmt.Sprintf("Bearer %s", testIapToken), r.Header.Get("Authorization"))
			require.Equal(t, testGhaToken, r.Header.Get(sherlockGhaOidcHeader))
			w.Header().Add("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockOkResponse{Ok: true})
		}))
		defer mockSherlockServer.Close()

		thelmaConfig, err := config.Load(config.WithTestDefaults(t), config.WithOverride("sherlock.addr", mockSherlockServer.URL))
		require.NoError(t, err)

		client, err := NewClient(testIapToken, func(options *Options) {
			options.ConfigSource = thelmaConfig
			options.GhaOidcTokenProvider = &credentials.MockTokenProvider{ReturnString: testGhaToken}
		})
		require.NoError(t, err)

		err = client.GetStatus()
		require.NoError(t, err)
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
	client, err := NewClient("fake", func(options *Options) {
		options.ConfigSource = suite.config
	})
	suite.Assert().NoError(err)
	envs, err := client.Environments()
	suite.Assert().NoError(err)

	suite.Assert().Equal(2, len(envs))
	suite.Assert().Equal("dev", envs[0].Name)
}

func (suite *sherlockClientSuite) TestFetchClusters() {
	client, err := NewClient("fake", func(options *Options) {
		options.ConfigSource = suite.config
	})
	suite.Assert().NoError(err)
	clusters, err := client.Clusters()
	suite.Assert().NoError(err)

	suite.Assert().Equal(2, len(clusters))
	suite.Assert().Equal("5.6.7.8", clusters[1].Address)
}

func (suite *sherlockClientSuite) TestFetchReleases() {
	client, err := NewClient("fake", func(options *Options) {
		options.ConfigSource = suite.config
	})
	suite.Assert().NoError(err)
	releases, err := client.Releases()
	suite.Assert().NoError(err)

	suite.Assert().Equal(3, len(releases))
	suite.Assert().Equal("sam", releases[0].Chart)
}

func (suite *sherlockClientSuite) TestFetchEnvironmentsError() {
	client, err := NewClient("fake", func(options *Options) {
		options.ConfigSource = suite.errConfig
	})
	suite.Assert().NoError(err)
	_, err = client.Environments()
	suite.Assert().Error(err)
}

func (suite *sherlockClientSuite) TestFetchClustersError() {
	client, err := NewClient("fake", func(options *Options) {
		options.ConfigSource = suite.errConfig
	})
	suite.Assert().NoError(err)
	_, err = client.Clusters()
	suite.Assert().Error(err)
}

func (suite *sherlockClientSuite) TestFetchReleasesError() {
	client, err := NewClient("fake", func(options *Options) {
		options.ConfigSource = suite.errConfig
	})
	suite.Assert().NoError(err)
	_, err = client.Releases()
	suite.Assert().Error(err)
}

func newMockSherlockServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/environments", mockEnvironmentsHandler())
	mux.HandleFunc("/api/clusters/v3", mockClustersHandler())
	mux.HandleFunc("/api/v2/chart-releases", mockChartReleasesHandler())
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
		_ = json.NewEncoder(w).Encode([]*models.SherlockClusterV3{
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
		_ = encoder.Encode([]*models.V2controllersChartRelease{
			{
				DestinationType:   "environment",
				AppVersionExact:   "0.2.1",
				Chart:             "sam",
				ChartVersionExact: "1.2.3",
				Name:              "sam-dev",
				Cluster:           "terra-dev",
				Environment:       "dev",
			},
			{
				DestinationType:   "cluster",
				AppVersionExact:   "2.2.1",
				Chart:             "grafana",
				ChartVersionExact: "0.0.3",
				Name:              "grafana-tools",
				Cluster:           "tools",
			},
			{
				DestinationType:   "cluster",
				AppVersionExact:   "1.0.5",
				Chart:             "argocd",
				ChartVersionExact: "0.3.1",
				Name:              "argocd-tools",
				Cluster:           "tools",
			},
		})

	}
}

func mockErroringHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

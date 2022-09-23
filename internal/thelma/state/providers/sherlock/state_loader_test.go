package sherlock

import (
	"fmt"
	"testing"

	"github.com/broadinstitute/sherlock/clients/go/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/suite"
)

func TestSherlockStateLoaderProviderSuite(t *testing.T) {
	suite.Run(t, new(sherlockStateLoaderSuite))
}

type sherlockStateLoaderSuite struct {
	suite.Suite
}

func (suite *sherlockStateLoaderSuite) TestStateLoading() {

	stateSource := mocks.NewStateLoader(suite.T())
	setStateExpectations(stateSource)

	thelmaHome := suite.T().TempDir()
	runner := shell.DefaultMockRunner()
	s := NewStateLoader(thelmaHome, runner, stateSource)
	state, err := s.Load()

	suite.Assert().NoError(err)
	suite.Assert().NotNil(state)

	_environments := state.Environments()
	_clusters := state.Clusters()

	// ensure the state object was actually populated with data
	suite.Assert().NotNil(_environments)
	suite.Assert().NotNil(_clusters)

	environments, err := _environments.All()
	suite.Assert().NoError(err)
	suite.Assert().Equal(2, len(environments))

	clusters, err := _clusters.All()
	suite.Assert().NoError(err)
	suite.Assert().Equal(2, len(clusters))

	suite.Assert().Equal(1, len(clusters[0].Releases()))
	suite.Assert().Equal(1, len(environments[1].Releases()))

	devEnv, err := _environments.Get("dev")
	suite.Assert().NoError(err)
	devEnvReleases := devEnv.Releases()
	suite.Assert().Equal("datarepo", devEnvReleases[0].Name())

	prodCluster, err := _clusters.Get("terra-prod")
	suite.Assert().NoError(err)
	prodClusterReleases := prodCluster.Releases()
	suite.Assert().Equal("sam", prodClusterReleases[0].Name())
}

func (suite *sherlockStateLoaderSuite) TestStateLoadingError() {
	stateSource := mocks.NewStateLoader(suite.T())
	errMsg := "this is an error from sherlock"
	stateSource.On("Clusters").Return(nil, fmt.Errorf(errMsg))

	thelmaHome := suite.T().TempDir()
	runner := shell.DefaultMockRunner()
	s := NewStateLoader(thelmaHome, runner, stateSource)
	state, err := s.Load()
	suite.Assert().Error(err)
	suite.Assert().ErrorContains(err, errMsg)
	suite.Assert().Nil(state)
}

func nullableBool(b bool) *bool {
	return &b
}

func nullableString(s string) *string {
	return &s
}

//nolint:govet // Ignore checks for unkeyed nested struct literals
func setStateExpectations(mock *mocks.StateLoader) {
	mock.On("Clusters").Return(
		sherlock.Clusters{
			sherlock.Cluster{
				&models.V2controllersCluster{
					Name:                "terra-dev",
					GoogleProject:       "dev-proj",
					Address:             "10.10.10.10",
					RequiresSuitability: nullableBool(false),
					Provider:            nullableString("google"),
				},
			},
			sherlock.Cluster{
				&models.V2controllersCluster{
					Name:                "terra-prod",
					GoogleProject:       "prod-proj",
					Address:             "10.10.10.11",
					RequiresSuitability: nullableBool(true),
					Provider:            nullableString("google"),
				},
			},
		}, nil,
	)

	mock.On("Environments").Return(
		sherlock.Environments{
			sherlock.Environment{
				&models.V2controllersEnvironment{
					Name:                "dev",
					Base:                "live",
					BaseDomain:          nullableString("dsde-dev.broadinstitute.org"),
					DefaultCluster:      "terra-dev",
					DefaultNamespace:    "terra-dev",
					Lifecycle:           nullableString("static"),
					RequiresSuitability: nullableBool(false),
				},
			},
			sherlock.Environment{
				&models.V2controllersEnvironment{
					Name:                "prod",
					Base:                "live",
					BaseDomain:          nullableString("dsde-prod.broadinstitute.org"),
					DefaultCluster:      "terra-prod",
					DefaultNamespace:    "terra-prod",
					Lifecycle:           nullableString("static"),
					RequiresSuitability: nullableBool(true),
				},
			},
		}, nil,
	)

	mock.On("ClusterReleases", "terra-dev").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					AppVersionExact:   "1.0.1",
					Chart:             "sam",
					ChartVersionExact: "0.43.0",
					Cluster:           "terra-dev",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
					Environment: "dev",
					Name:        "sam-dev",
					Namespace:   "terra-dev",
					HelmfileRef: nullableString("asdf"),
				},
			},
		}, nil,
	)

	mock.On("ClusterReleases", "terra-prod").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					AppVersionExact:   "1.0.0",
					Chart:             "sam",
					ChartVersionExact: "0.42.0",
					Cluster:           "terra-prod",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
					Environment: "prod",
					Name:        "sam-prod",
					Namespace:   "terra-prod",
				},
			},
		}, nil,
	)

	mock.On("EnvironmentReleases", "dev").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					AppVersionExact:   "0.160.0",
					Chart:             "datarepo",
					ChartVersionExact: "0.33.0",
					Cluster:           "terra-dev",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
					Environment: "dev",
					Name:        "datarepo-dev",
					Namespace:   "terra-dev",
					HelmfileRef: nullableString("oisgff"),
				},
			},
		}, nil,
	)

	mock.On("EnvironmentReleases", "prod").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					AppVersionExact:   "0.156.0",
					Chart:             "datarepo",
					ChartVersionExact: "0.32.0",
					Cluster:           "terra-prod",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
					Environment: "prod",
					Name:        "datarepo-prod",
					Namespace:   "terra-prod",
					HelmfileRef: nullableString("wlekjerw"),
				},
			},
		}, nil,
	)

}

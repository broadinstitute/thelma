package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"testing"

	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/stretchr/testify/suite"
)

func TestSherlockStateLoaderProviderSuite(t *testing.T) {
	suite.Run(t, new(sherlockStateLoaderSuite))
}

type sherlockStateLoaderSuite struct {
	suite.Suite
}

func (suite *sherlockStateLoaderSuite) TestStateLoading() {

	stateSource := mocks.NewClient(suite.T())
	setStateExpectations(stateSource)

	thelmaHome := suite.T().TempDir()
	s := NewStateLoader(thelmaHome, stateSource)
	state, err := s.Load()

	suite.Assert().NoError(err)
	suite.Assert().NotNil(state)

	_environments := state.Environments()
	_clusters := state.Clusters()
	_releases := state.Releases()

	// ensure the state object was actually populated with data
	suite.Assert().NotNil(_environments)
	suite.Assert().NotNil(_clusters)

	environments, err := _environments.All()
	suite.Assert().NoError(err)
	suite.Assert().Equal(4, len(environments))

	clusters, err := _clusters.All()
	suite.Assert().NoError(err)
	suite.Assert().Equal(3, len(clusters))

	devCluster, err := _clusters.Get("terra-dev")
	suite.Assert().NoError(err)
	suite.Assert().Len(devCluster.Releases(), 1)

	devEnv, err := _environments.Get("dev")
	suite.Assert().NoError(err)
	devEnvReleases := devEnv.Releases()
	suite.Assert().Len(devEnvReleases, 1)
	suite.Assert().Equal("datarepo", devEnvReleases[0].Name())

	prodCluster, err := _clusters.Get("terra-prod")
	suite.Assert().NoError(err)
	prodClusterReleases := prodCluster.Releases()
	suite.Assert().Equal("sam-prod", prodClusterReleases[0].Name())
	suite.Assert().Equal("12.13.14", prodClusterReleases[0].AppVersion())

	onlineBeeEnv, err := _environments.Get("bee-online")
	suite.Assert().NoError(err)
	suite.Assert().False(onlineBeeEnv.Offline())
	for _, release := range onlineBeeEnv.Releases() {
		suite.Assert().Len(release.HelmfileOverlays(), 0)
	}

	offlineBeeEnv, err := _environments.Get("bee-offline")
	suite.Assert().NoError(err)
	suite.Assert().True(offlineBeeEnv.Offline())
	for _, release := range offlineBeeEnv.Releases() {
		suite.Assert().Equal([]string{"offline"}, release.HelmfileOverlays())
	}

	allReleases, err := _releases.All()
	suite.Assert().NoError(err)
	suite.Assert().Equal(6, len(allReleases))
	for _, r := range allReleases {
		if r.IsAppRelease() {
			r.(terra.AppRelease).AppVersion()
		} else {
			r.(terra.ClusterRelease).IsClusterRelease()
		}
	}

	// Calling Load() is cached
	stateSource.AssertNumberOfCalls(suite.T(), "Releases", 1)
	_, err = s.Load()
	suite.Assert().NoError(err)
	stateSource.AssertNumberOfCalls(suite.T(), "Releases", 1)

	// Calling Reload() is not
	_, err = s.Reload()
	suite.Assert().NoError(err)
	stateSource.AssertNumberOfCalls(suite.T(), "Releases", 2)
	_, err = s.Load()
	suite.Assert().NoError(err)
	stateSource.AssertNumberOfCalls(suite.T(), "Releases", 2)
}

func (suite *sherlockStateLoaderSuite) TestStateLoadingError() {
	stateSource := mocks.NewClient(suite.T())
	errMsg := "this is an error from sherlock"
	stateSource.On("Clusters").Return(nil, errors.Errorf(errMsg))

	thelmaHome := suite.T().TempDir()
	s := NewStateLoader(thelmaHome, stateSource)
	state, err := s.Load()
	suite.Assert().Error(err)
	suite.Assert().ErrorContains(err, errMsg)
	suite.Assert().Nil(state)
}

//nolint:govet // Ignore checks for unkeyed nested struct literals
func setStateExpectations(mock *mocks.Client) {
	mock.On("Clusters").Return(
		sherlock.Clusters{
			sherlock.Cluster{
				&models.SherlockClusterV3{
					Name:                "terra-dev",
					GoogleProject:       "dev-proj",
					Address:             "10.10.10.10",
					RequiresSuitability: utils.Nullable(false),
					Provider:            utils.Nullable("google"),
					Location:            utils.Nullable("us-central1-a"),
					HelmfileRef:         utils.Nullable("HEAD"),
				},
			},
			sherlock.Cluster{
				&models.SherlockClusterV3{
					Name:                "terra-prod",
					GoogleProject:       "prod-proj",
					Address:             "10.10.10.11",
					RequiresSuitability: utils.Nullable(true),
					Provider:            utils.Nullable("google"),
					Location:            utils.Nullable("us-central-1"),
					HelmfileRef:         utils.Nullable("HEAD"),
				},
			},
			sherlock.Cluster{
				&models.SherlockClusterV3{
					Name:                "terra-qa-bees",
					GoogleProject:       "broad-dsde-qa",
					Address:             "10.10.10.12",
					RequiresSuitability: utils.Nullable(false),
					Provider:            utils.Nullable("google"),
					Location:            utils.Nullable("us-central-1"),
					HelmfileRef:         utils.Nullable("HEAD"),
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
					BaseDomain:          utils.Nullable("dsde-dev.broadinstitute.org"),
					DefaultCluster:      "terra-dev",
					DefaultNamespace:    "terra-dev",
					Lifecycle:           utils.Nullable("static"),
					RequiresSuitability: utils.Nullable(false),
					NamePrefixesDomain:  utils.Nullable(true),
					HelmfileRef:         utils.Nullable("HEAD"),
					PreventDeletion:     utils.Nullable(false),
				},
			},
			sherlock.Environment{
				&models.V2controllersEnvironment{
					Name:                "prod",
					Base:                "live",
					BaseDomain:          utils.Nullable("dsde-prod.broadinstitute.org"),
					DefaultCluster:      "terra-prod",
					DefaultNamespace:    "terra-prod",
					Lifecycle:           utils.Nullable("static"),
					RequiresSuitability: utils.Nullable(true),
					NamePrefixesDomain:  utils.Nullable(false),
					HelmfileRef:         utils.Nullable("HEAD"),
					PreventDeletion:     utils.Nullable(false),
				},
			},
			sherlock.Environment{
				&models.V2controllersEnvironment{
					Name:                "bee-online",
					Base:                "bee",
					BaseDomain:          utils.Nullable("bee.envs-terra.bio"),
					DefaultCluster:      "terra-qa-bees",
					DefaultNamespace:    "terra-bee-online",
					Lifecycle:           utils.Nullable("dynamic"),
					RequiresSuitability: utils.Nullable(false),
					NamePrefixesDomain:  utils.Nullable(true),
					HelmfileRef:         utils.Nullable("HEAD"),
					PreventDeletion:     utils.Nullable(false),
				},
			},
			sherlock.Environment{
				&models.V2controllersEnvironment{
					Name:                "bee-offline",
					Base:                "bee",
					BaseDomain:          utils.Nullable("bee.envs-terra.bio"),
					DefaultCluster:      "terra-qa-bees",
					DefaultNamespace:    "terra-bee-offline",
					Lifecycle:           utils.Nullable("dynamic"),
					RequiresSuitability: utils.Nullable(false),
					NamePrefixesDomain:  utils.Nullable(true),
					HelmfileRef:         utils.Nullable("HEAD"),
					PreventDeletion:     utils.Nullable(false),
					Offline:             utils.Nullable(true),
				},
			},
		}, nil,
	)

	mock.On("Releases").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					DestinationType:   "cluster",
					AppVersionExact:   "1.0.1",
					Chart:             "sam",
					ChartVersionExact: "0.43.0",
					Cluster:           "terra-dev",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: utils.Nullable(""),
					},
					Environment:         "dev",
					Name:                "sam-dev",
					Namespace:           "terra-dev",
					HelmfileRef:         utils.Nullable("wlekjerw"),
					FirecloudDevelopRef: "",
				},
			},
			sherlock.Release{
				&models.V2controllersChartRelease{
					DestinationType:   "cluster",
					AppVersionExact:   "12.13.14",
					Chart:             "sam",
					ChartVersionExact: "0.42.0",
					Cluster:           "terra-prod",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: utils.Nullable(""),
					},
					Environment:         "prod",
					Name:                "sam-prod",
					Namespace:           "terra-prod",
					HelmfileRef:         utils.Nullable("wlekjerw"),
					FirecloudDevelopRef: "",
				},
			},
			sherlock.Release{
				&models.V2controllersChartRelease{
					DestinationType:   "environment",
					AppVersionExact:   "0.160.0",
					Chart:             "datarepo",
					ChartVersionExact: "0.33.0",
					Cluster:           "terra-dev",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: utils.Nullable(""),
					},
					Environment:         "dev",
					Name:                "datarepo-dev",
					Namespace:           "terra-dev",
					HelmfileRef:         utils.Nullable("oisgff"),
					FirecloudDevelopRef: "",
				},
			},
			sherlock.Release{
				&models.V2controllersChartRelease{
					DestinationType:   "environment",
					AppVersionExact:   "0.156.0",
					Chart:             "datarepo",
					ChartVersionExact: "0.32.0",
					Cluster:           "terra-prod",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: utils.Nullable(""),
					},
					Environment:         "prod",
					Name:                "datarepo-prod",
					Namespace:           "terra-prod",
					HelmfileRef:         utils.Nullable("wlekjerw"),
					FirecloudDevelopRef: "",
				},
			},
			sherlock.Release{
				&models.V2controllersChartRelease{
					DestinationType:   "environment",
					AppVersionExact:   "1.0.1",
					Chart:             "sam",
					ChartVersionExact: "0.43.0",
					Cluster:           "terra-qa-bees",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: utils.Nullable(""),
					},
					Environment:         "bee-online",
					Name:                "sam-bee-online",
					Namespace:           "terra-bee-online",
					HelmfileRef:         utils.Nullable("wlekjerw"),
					FirecloudDevelopRef: "",
				},
			},
			sherlock.Release{
				&models.V2controllersChartRelease{
					DestinationType:   "environment",
					AppVersionExact:   "1.0.1",
					Chart:             "sam",
					ChartVersionExact: "0.43.0",
					Cluster:           "terra-qa-bees",
					ChartInfo: &models.V2controllersChart{
						ChartRepo: utils.Nullable(""),
					},
					Environment:         "bee-offline",
					Name:                "sam-bee-offline",
					Namespace:           "terra-bee-offline",
					HelmfileRef:         utils.Nullable("wlekjerw"),
					FirecloudDevelopRef: "",
				},
			},
		}, nil,
	)

}

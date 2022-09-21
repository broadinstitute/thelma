package sherlock

import (
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

//nolint:govet // Ignore checks for unkeyed nested struct literals
func (suite *sherlockStateLoaderSuite) TestStateLoading() {

	stateSource := mocks.NewStateLoader(suite.T())
	stateSource.On("Clusters").Return(
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

	stateSource.On("Environments").Return(
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

	stateSource.On("ClusterReleases", "terra-dev").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
				},
			},
		}, nil,
	)

	stateSource.On("ClusterReleases", "terra-prod").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
				},
			},
		}, nil,
	)

	stateSource.On("EnvironmentReleases", "dev").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
				},
			},
		}, nil,
	)

	stateSource.On("EnvironmentReleases", "prod").Return(
		sherlock.Releases{
			sherlock.Release{
				&models.V2controllersChartRelease{
					ChartInfo: &models.V2controllersChart{
						ChartRepo: nullableString(""),
					},
				},
			},
		}, nil,
	)

	thelmaHome := suite.T().TempDir()
	runner := shell.DefaultMockRunner()
	s := NewStateLoader(thelmaHome, runner, stateSource)
	state, err := s.Load()
	suite.Assert().NoError(err)
	suite.Assert().NotNil(state)
}

func nullableBool(b bool) *bool {
	return &b
}

func nullableString(s string) *string {
	return &s
}

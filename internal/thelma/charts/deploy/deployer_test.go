package deploy

import (
	deploymocks "github.com/broadinstitute/thelma/internal/thelma/charts/deploy/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/charts/releaser"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	sherlockmocks "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sync"
	syncmocks "github.com/broadinstitute/thelma/internal/thelma/ops/sync/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DeployerSuite struct {
	suite.Suite
	mockConfigLoader *deploymocks.ConfigLoader
	mockStateLoader  *statemocks.StateLoader
	mockSherlock     *sherlockmocks.ChartVersionUpdater
	mockSync         *syncmocks.Sync
}

func (suite *DeployerSuite) SetupTest() {
	suite.mockConfigLoader = deploymocks.NewConfigLoader(suite.T())
	suite.mockStateLoader = statemocks.NewStateLoader(suite.T())
	suite.mockSherlock = sherlockmocks.NewChartVersionUpdater(suite.T())
	suite.mockSync = syncmocks.NewSync(suite.T())
}

func (suite *DeployerSuite) TestUpdatesAndSyncOneChartRelease() {
	releases := mockReleases([]string{"agora-dev"})

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("agora").
		Return(releases, nil)

	suite.mockSherlock.EXPECT().
		UpdateForNewChartVersion(
			"agora",
			"1.2.4",
			"1.2.3",
			"a change description",
			[]string{"agora-dev"},
		).Return(nil)

	suite.expectStateReloadAndReturnReleases(releases)

	suite.mockSync.EXPECT().
		Sync(releases, maxParallelSync).
		Return(nil, nil)

	_deployer := suite.newDeployer(Options{
		DryRun:            false,
		IgnoreSyncFailure: false,
	})

	err := _deployer.Deploy(map[string]releaser.VersionPair{
		"agora": {
			PriorVersion: "1.2.3",
			NewVersion:   "1.2.4",
		},
	}, "a change description")

	require.NoError(suite.T(), err)
}

func (suite *DeployerSuite) TestUpdatesAndSyncMultipleChartsAndReleases() {
	agoraDev := mockRelease("agora-dev")
	samStaging := mockRelease("sam-alpha")
	yaleDev := mockRelease("yale-terra-dev")
	yaleBees := mockRelease("yale-terra-qa-bees")

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("agora").
		Return([]terra.Release{agoraDev}, nil)

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("sam").
		Return([]terra.Release{samStaging}, nil)

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("yale").
		Return([]terra.Release{yaleDev, yaleBees}, nil)

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("httpd-proxy").
		Return([]terra.Release{}, nil)

	suite.mockSherlock.EXPECT().
		UpdateForNewChartVersion(
			"agora",
			"0.0.5",
			"0.0.2",
			"my multi-chart change",
			[]string{"agora-dev"},
		).Return(nil)

	suite.mockSherlock.EXPECT().
		UpdateForNewChartVersion(
			"sam",
			"10.11.12",
			"10.11.11",
			"my multi-chart change",
			[]string{"sam-alpha"},
		).Return(nil)

	suite.mockSherlock.EXPECT().
		UpdateForNewChartVersion(
			"yale",
			"8.0.0",
			"7.9.9",
			"my multi-chart change",
			[]string{"yale-terra-dev", "yale-terra-qa-bees"},
		).Return(nil)

	suite.expectStateReloadAndReturnReleases([]terra.Release{
		agoraDev,
		samStaging,
		yaleDev,
		yaleBees,
		mockRelease("agora-prod"), // throw in an extra just to verify it's ignored
	})

	suite.mockSync.EXPECT().
		Sync(mock.Anything, maxParallelSync).
		Run(func(releases []terra.Release, _maxParallel int, _opts ...argocd.SyncOption) {
			// release order is unpredictable, so we sort before asserting
			assert.ElementsMatch(suite.T(), releases, []terra.Release{
				agoraDev,
				samStaging,
				yaleDev,
				yaleBees,
			})
		}).
		Return(nil, nil)

	_deployer := suite.newDeployer(Options{
		DryRun:            false,
		IgnoreSyncFailure: false,
	})

	err := _deployer.Deploy(map[string]releaser.VersionPair{
		"agora": {
			NewVersion:   "0.0.5",
			PriorVersion: "0.0.2",
		},
		"sam": {
			NewVersion:   "10.11.12",
			PriorVersion: "10.11.11",
		},
		"yale": {
			NewVersion:   "8.0.0",
			PriorVersion: "7.9.9",
		},
		"httpd-proxy": {
			NewVersion:   "1.2.3",
			PriorVersion: "0.2.1",
		},
	}, "my multi-chart change")

	require.NoError(suite.T(), err)
}

func (suite *DeployerSuite) TestDryRunDoesNotUpdateOrSync() {
	releases := mockReleases([]string{"agora-dev"})

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("agora").
		Return(releases, nil)

	// no updates or syncs should be called

	_deployer := suite.newDeployer(Options{
		DryRun:            true,
		IgnoreSyncFailure: false,
	})

	suite.expectStateReloadAndReturnReleases(releases)

	err := _deployer.Deploy(map[string]releaser.VersionPair{
		"agora": {
			PriorVersion: "1.2.3",
			NewVersion:   "1.2.4",
		},
	}, "a change description")

	require.NoError(suite.T(), err)
}

func (suite *DeployerSuite) TestSyncFailuresAreReportedIfIgnoreIsFalse() {
	releases := mockReleases([]string{"agora-dev"})

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("agora").
		Return(releases, nil)

	suite.mockSherlock.EXPECT().
		UpdateForNewChartVersion(
			"agora",
			"1.2.4",
			"1.2.3",
			"a change description",
			[]string{"agora-dev"},
		).Return(nil)

	suite.expectStateReloadAndReturnReleases(releases)

	suite.mockSync.EXPECT().
		Sync(releases, maxParallelSync).
		Return(nil, errors.Errorf("oops, the sync failed"))

	_deployer := suite.newDeployer(Options{
		DryRun:            false,
		IgnoreSyncFailure: false,
	})

	err := _deployer.Deploy(map[string]releaser.VersionPair{
		"agora": {
			PriorVersion: "1.2.3",
			NewVersion:   "1.2.4",
		},
	}, "a change description")

	assert.Error(suite.T(), err)
	assert.ErrorContains(suite.T(), err, "oops, the sync failed")
}

func (suite *DeployerSuite) TestSyncFailuresAreSuppredIfIgnoreIsTrue() {
	releases := mockReleases([]string{"agora-dev"})

	suite.mockConfigLoader.EXPECT().
		FindReleasesToUpdate("agora").
		Return(releases, nil)

	suite.mockSherlock.EXPECT().
		UpdateForNewChartVersion(
			"agora",
			"1.2.4",
			"1.2.3",
			"a change description",
			[]string{"agora-dev"},
		).Return(nil)

	suite.expectStateReloadAndReturnReleases(releases)

	suite.mockSync.EXPECT().
		Sync(releases, maxParallelSync).
		Return(nil, errors.Errorf("oops, the sync failed"))

	_deployer := suite.newDeployer(Options{
		DryRun:            false,
		IgnoreSyncFailure: true,
	})

	err := _deployer.Deploy(map[string]releaser.VersionPair{
		"agora": {
			PriorVersion: "1.2.3",
			NewVersion:   "1.2.4",
		},
	}, "a change description")

	require.NoError(suite.T(), err)
}

func TestDeployerSuite(t *testing.T) {
	suite.Run(t, new(DeployerSuite))
}

func (suite *DeployerSuite) newDeployer(opts Options) Deployer {
	updater := DeployedVersionUpdater{
		SherlockUpdaters: []sherlock.ChartVersionUpdater{suite.mockSherlock},
	}

	syncFactory := func() (sync.Sync, error) {
		return suite.mockSync, nil
	}

	return newForTesting(suite.mockConfigLoader, updater, suite.mockStateLoader, syncFactory, opts)
}

func (suite *DeployerSuite) expectStateReloadAndReturnReleases(releases []terra.Release) {
	state := &statemocks.State{}
	releaseInterface := &statemocks.Releases{}

	state.EXPECT().Releases().Return(releaseInterface)
	releaseInterface.EXPECT().All().Return(releases, nil)
	suite.mockStateLoader.EXPECT().Reload().Return(state, nil)
}

func mockReleases(fullNames []string) []terra.Release {
	var releases []terra.Release
	for _, n := range fullNames {
		releases = append(releases, mockRelease(n))
	}
	return releases
}

func mockRelease(fullName string) terra.Release {
	r := &statemocks.Release{}
	r.EXPECT().FullName().Return(fullName)
	return r
}

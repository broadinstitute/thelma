package releaser

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/sync"
	syncmocks "github.com/broadinstitute/thelma/internal/thelma/ops/sync/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PostUpdateSyncerSuite struct {
	suite.Suite
	mockSync    *syncmocks.Sync
	syncFactory func() (sync.Sync, error)
	state       *statemocks.State
}

func (suite *PostUpdateSyncerSuite) SetupTest() {
	statefixture, err := statefixtures.LoadFixtureFromFile("testdata/statefixture.yaml")
	require.NoError(suite.T(), err)

	suite.mockSync = syncmocks.NewSync(suite.T())
	suite.syncFactory = func() (sync.Sync, error) {
		return suite.mockSync, nil
	}
	suite.state = statefixture.Mocks().State
}

func (suite *PostUpdateSyncerSuite) TestDoesNotSyncIfDryRun() {
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, Options{DryRun: true, IgnoreSyncFailure: false})
	err := syncer.Sync([]string{"agora-dev", "sam-dev"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestDoesNotSyncIfNoReleases() {
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, Options{DryRun: true, IgnoreSyncFailure: false})
	err := syncer.Sync([]string{})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestIgnoresReleasesMissingFromState() {
	suite.expectSyncFor("agora-dev", "sam-dev", "yale-terra-dev").Return(nil, nil)
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, Options{DryRun: false, IgnoreSyncFailure: false})
	err := syncer.Sync([]string{"agora-dev", "sam-dev", "this-release-does-not-exist-in-state", "yale-terra-dev"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestNoSyncCallIfNoMatchingReleasesFound() {
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, Options{DryRun: false, IgnoreSyncFailure: false})
	err := syncer.Sync([]string{"not-in-state-1", "not-in-state-2"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestSyncsAllReleases() {
	suite.expectSyncFor("agora-dev", "sam-dev", "yale-terra-dev").Return(nil, nil)
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, Options{DryRun: false, IgnoreSyncFailure: false})
	err := syncer.Sync([]string{"agora-dev", "sam-dev", "yale-terra-dev"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestReturnsErrorIfSyncFailsAndIgnoreSyncFailureIsFalse() {
	suite.expectSyncFor("agora-dev", "sam-dev", "yale-terra-dev").Return(nil, errors.Errorf("sync failed!"))
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, Options{DryRun: false, IgnoreSyncFailure: false})
	err := syncer.Sync([]string{"agora-dev", "sam-dev", "yale-terra-dev"})
	assert.Error(suite.T(), err)
	assert.ErrorContains(suite.T(), err, "sync failed!")
}

func (suite *PostUpdateSyncerSuite) TestDoesNotReturnErrorIfSyncFailsAndIgnoreSyncFailureIsFalse() {
	suite.expectSyncFor("agora-dev", "sam-dev", "yale-terra-dev").Return(nil, errors.Errorf("sync failed!"))
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, Options{DryRun: false, IgnoreSyncFailure: true})
	err := syncer.Sync([]string{"agora-dev", "sam-dev", "yale-terra-dev"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) expectSyncFor(chartReleaseNames ...string) *syncmocks.Sync_Sync_Call {
	// this is annoyingly complicated because mock state does not return releases in any particular order...
	// so we write a mock interceptor to pull the names off the releases that are passed in and then compare them
	return suite.mockSync.EXPECT().Sync(
		mock.MatchedBy(func(releases []terra.Release) bool {
			var names []string
			for _, r := range releases {
				names = append(names, r.FullName())
			}
			assert.ElementsMatch(suite.T(), chartReleaseNames, names)
			return true
		}),
		maxParallelSync,
	)
}

func TestPostUpdateSyncerSuite(t *testing.T) {
	suite.Run(t, new(PostUpdateSyncerSuite))
}

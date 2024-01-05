package releaser

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/sync"
	syncmocks "github.com/broadinstitute/thelma/internal/thelma/ops/sync/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
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
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, true)
	err := syncer.Sync([]string{"agora-dev", "sam-dev"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestDoesNotSyncIfNoReleases() {
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, false)
	err := syncer.Sync([]string{})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestIgnoresReleasesMissingFromState() {
	suite.assertSyncCalledForReleases("agora-dev", "sam-dev", "yale-terra-dev")
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, false)
	err := syncer.Sync([]string{"agora-dev", "sam-dev", "this-release-does-not-exist-in-state", "yale-terra-dev"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestNoSyncCallIfNoMatchingReleasesFound() {
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, false)
	err := syncer.Sync([]string{"not-in-state-1", "not-in-state-2"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) TestSyncsAllReleases() {
	suite.assertSyncCalledForReleases("agora-dev", "sam-dev", "yale-terra-dev")
	syncer := NewPostUpdateSyncer(suite.syncFactory, suite.state, false)
	err := syncer.Sync([]string{"agora-dev", "sam-dev", "yale-terra-dev"})
	require.NoError(suite.T(), err)
}

func (suite *PostUpdateSyncerSuite) assertSyncCalledForReleases(chartReleaseNames ...string) {
	// this is annoyingly complicated because mock state does not return releases in any particular order...
	// so we write a mock interceptor to pull the names off the releases that are passed in and then compare them
	suite.mockSync.EXPECT().Sync(
		mock.MatchedBy(func(releases []terra.Release) bool {
			var names []string
			for _, r := range releases {
				names = append(names, r.FullName())
			}
			assert.ElementsMatch(suite.T(), chartReleaseNames, names)
			return true
		}),
		maxParallelSync,
	).Return(nil, nil) // we don't bother returning a fake status map because it is ignored
}

func TestPostUpdateSyncerSuite(t *testing.T) {
	suite.Run(t, new(PostUpdateSyncerSuite))
}

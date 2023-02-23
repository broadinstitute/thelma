package installer

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releasebucket"
	releasebucketmocks "github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releasebucket/mocks"
	releasesmocks "github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const keepReleases = 8

type InstallerSuite struct {
	suite.Suite
	bucket    *releasebucketmocks.ReleaseBucket
	dir       *releasesmocks.Dir
	installer Installer
}

func TestInstaller(t *testing.T) {
	suite.Run(t, new(InstallerSuite))
}

func (suite *InstallerSuite) SetupTest() {
	t := suite.T()

	bucket := releasebucketmocks.NewReleaseBucket(t)
	dir := releasesmocks.NewDir(t)
	_installer := New(dir, bucket, func(options *Options) {
		options.KeepReleases = keepReleases
	})

	suite.bucket = bucket
	suite.dir = dir
	suite.installer = _installer
}

func (suite *InstallerSuite) Test_ResolveVersions_UpdatedNeeded() {
	suite.dir.EXPECT().CurrentVersion().Return("v1.0.0", nil)
	suite.bucket.EXPECT().ResolveTagOrVersion("latest").Return("v1.2.3", nil)

	resolved, err := suite.installer.ResolveVersions("latest")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v1.0.0", resolved.CurrentVersion)
	assert.Equal(suite.T(), "v1.2.3", resolved.TargetVersion)
	assert.Equal(suite.T(), "latest", resolved.VersionAlias)
	assert.True(suite.T(), resolved.UpdateNeeded())
}

func (suite *InstallerSuite) Test_ResolveVersions_UpdatedNotNeeded() {
	suite.dir.EXPECT().CurrentVersion().Return("v4.5.6", nil)
	suite.bucket.EXPECT().ResolveTagOrVersion("unstable").Return("v4.5.6", nil)

	resolved, err := suite.installer.ResolveVersions("unstable")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v4.5.6", resolved.CurrentVersion)
	assert.Equal(suite.T(), "v4.5.6", resolved.TargetVersion)
	assert.Equal(suite.T(), "unstable", resolved.VersionAlias)
	assert.False(suite.T(), resolved.UpdateNeeded())
}

func (suite *InstallerSuite) Test_UpdateThelma() {
	suite.dir.EXPECT().CurrentVersion().Return("v1.0.0", nil)
	suite.bucket.EXPECT().ResolveTagOrVersion("latest").Return("v2.0.0", nil)

	suite.dir.EXPECT().WithInstallerLock(mock.Anything).Run(func(fn func() error) {
		fnErr := fn()
		require.NoError(suite.T(), fnErr)
	}).Return(nil)

	tmpdir := suite.T().TempDir()
	suite.bucket.EXPECT().DownloadAndUnpack(mock.MatchedBy(func(archive releasebucket.Archive) bool {
		require.Equal(suite.T(), "v2.0.0", archive.Version())
		return true
	})).Return(tmpdir, nil)

	suite.dir.EXPECT().CopyUnpackedArchive(tmpdir).Return(nil)

	suite.dir.EXPECT().UpdateCurrentReleaseSymlink("v2.0.0").Return(nil)

	suite.dir.EXPECT().CleanupOldReleases(keepReleases).Return(nil)

	require.NoError(suite.T(), suite.installer.UpdateThelma("latest"))
}

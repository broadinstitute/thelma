package releases

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path"
	"sort"
	"testing"
)

type DirSuite struct {
	suite.Suite
	dirPath string
	dir     Dir
}

func TestDir(t *testing.T) {
	suite.Run(t, new(DirSuite))
}

func (suite *DirSuite) SetupTest() {
	t := suite.T()

	testConfig, err := config.NewTestConfig(t)
	require.NoError(t, err)

	_scratch, err := scratch.NewScratch(testConfig)
	require.NoError(t, err)

	dirPath := t.TempDir()

	dir := NewDir(dirPath, _scratch)

	suite.dirPath = dirPath
	suite.dir = dir
}

func TestCurrentReleaseSymlink(t *testing.T) {
	rootDir := t.TempDir()
	fakeRoot := root.NewAt(rootDir)
	assert.Equal(t, path.Join(rootDir, "releases", "current"), CurrentReleaseSymlink(fakeRoot))
}

func (suite *DirSuite) TestCurrentVersion() {
	_, err := suite.dir.CurrentVersion()
	require.Error(suite.T(), err)

	suite.CreateReleaseDir("v1.2.3")
	suite.SetCurrentSymlink("v1.2.3")

	version, err := suite.dir.CurrentVersion()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v1.2.3", version)
}

func (suite *DirSuite) TestCurrentVersionMatches() {
	matches := suite.dir.CurrentVersionMatches("v4.5.6")
	assert.False(suite.T(), matches)

	suite.CreateReleaseDir("v4.5.6")
	suite.SetCurrentSymlink("v4.5.6")

	matches = suite.dir.CurrentVersionMatches("v4.5.6")
	assert.True(suite.T(), matches)

	matches = suite.dir.CurrentVersionMatches("v1.2.5")
	assert.False(suite.T(), matches)
}

func (suite *DirSuite) TestUpdateCurrentReleaseSymlink_NoSymlink() {
	suite.CreateReleaseDir("v7.8.9")
	err := suite.dir.UpdateCurrentReleaseSymlink("v7.8.9")
	require.NoError(suite.T(), err)

	suite.AssertCurrentIsSymlink()
	suite.AssertCurrentPointsTo("v7.8.9")
}

func (suite *DirSuite) TestUpdateCurrentReleaseSymlink_SymlinkExists() {
	suite.CreateReleaseDir("v1.2.3")
	suite.CreateReleaseDir("v4.5.6")
	suite.CreateReleaseDir("v7.8.9")
	suite.SetCurrentSymlink("v1.2.3")

	err := suite.dir.UpdateCurrentReleaseSymlink("v7.8.9")
	require.NoError(suite.T(), err)

	suite.AssertCurrentIsSymlink()
	suite.AssertCurrentPointsTo("v7.8.9")
}

func (suite *DirSuite) TestCopyUnpackedArchive() {
	unpackDir := suite.CreateFakeUnpackedReleaseArchive("v4.5.6")

	err := suite.dir.CopyUnpackedArchive(unpackDir)
	require.NoError(suite.T(), err)

	assert.DirExists(suite.T(), suite.ReleaseDir("v4.5.6"))
	assert.FileExists(suite.T(), path.Join(suite.ReleaseDir("v4.5.6"), "build.json"))
}

func (suite *DirSuite) TestWithInstallerLock() {
	var count int

	err := suite.dir.WithInstallerLock(func() error {
		count++
		return errors.Errorf("fake error")
	})

	require.Error(suite.T(), err)
	assert.ErrorContains(suite.T(), err, "fake error")
	assert.Equal(suite.T(), 1, count)

	content, err := os.ReadFile(path.Join(suite.dirPath, lockFile))
	require.NoError(suite.T(), err)
	pid := os.Getpid()
	assert.Equal(suite.T(), fmt.Sprintf("%d", pid), string(content))
}

func (suite *DirSuite) TestCleanupOldReleases() {
	testCases := []struct {
		name            string
		localVersions   []string
		currentVersion  string
		keepReleases    int
		expectRemaining []string
	}{
		{
			name: "empty case",
		},
		{
			name:            "current version should not be deleted even if keep == 0",
			localVersions:   []string{"v1.2.3"},
			currentVersion:  "v1.2.3",
			keepReleases:    0,
			expectRemaining: []string{"v1.2.3"},
		},
		{
			name:            "5 versions, all should be deleted if keep == 0 and no current version",
			localVersions:   []string{"v1.2.3", "v1.2.4", "v1.2.5", "v1.2.6", "v1.2.7"},
			currentVersion:  "",
			keepReleases:    0,
			expectRemaining: []string{},
		},
		{
			name:            "3 versions, keep releases = 1, should delete 1 release",
			localVersions:   []string{"v1.2.3", "v1.2.4", "v1.2.5"},
			currentVersion:  "v1.2.5",
			keepReleases:    1,
			expectRemaining: []string{"v1.2.4", "v1.2.5"},
		},
		{
			name:            "should skip current even if not most recent release",
			localVersions:   []string{"v1.2.3", "v1.2.4", "v1.2.5", "v1.2.6", "v1.2.7"},
			currentVersion:  "v1.2.4",
			keepReleases:    2,
			expectRemaining: []string{"v1.2.4", "v1.2.6", "v1.2.7"},
		},
		{
			name:            "5 versions, keep releases = 4, should keep all",
			localVersions:   []string{"v1.2.3", "v1.2.4", "v1.2.5", "v1.2.6", "v1.2.7"},
			currentVersion:  "v1.2.3",
			keepReleases:    4,
			expectRemaining: []string{"v1.2.3", "v1.2.4", "v1.2.5", "v1.2.6", "v1.2.7"},
		},
		{
			name:            "should ignore invalid dirs",
			localVersions:   []string{"v1.2.3", "wtf", "ignoreme", "v1.2.6", "v1.2.7", "v1.2.8"},
			currentVersion:  "v1.2.6",
			keepReleases:    2,
			expectRemaining: []string{"wtf", "ignoreme", "v1.2.6", "v1.2.7", "v1.2.8"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			require.NoError(suite.T(), os.RemoveAll(suite.dirPath))
			require.NoError(suite.T(), os.MkdirAll(suite.dirPath, 0755))

			for _, version := range tc.localVersions {
				suite.CreateReleaseDir(version)
			}
			if tc.currentVersion != "" {
				suite.SetCurrentSymlink(tc.currentVersion)
			}
			require.NoError(suite.T(), suite.dir.CleanupOldReleases(tc.keepReleases))

			var remainingDirs []string
			entries, err := os.ReadDir(suite.dirPath)
			require.NoError(suite.T(), err)

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				remainingDirs = append(remainingDirs, entry.Name())
			}

			if tc.currentVersion != "" {
				// make sure cleanup did not remove the current symlink
				assert.True(suite.T(), suite.dir.CurrentVersionMatches(tc.currentVersion))
			}

			if len(tc.expectRemaining) == 0 {
				// separate check for len 0 to avoid nil/empty slice comparison failing assert.Equal
				assert.Empty(suite.T(), remainingDirs)
				return
			}

			sort.Strings(tc.expectRemaining)
			sort.Strings(remainingDirs)
			assert.Equal(suite.T(), tc.expectRemaining, remainingDirs)
		})
	}
}

func (suite *DirSuite) CreateFakeUnpackedReleaseArchive(version string) string {
	dir := suite.T().TempDir()
	manifestPath := path.Join(dir, "build.json")

	manifestJson, err := json.Marshal(map[string]string{"version": version})
	require.NoError(suite.T(), err)

	err = os.WriteFile(manifestPath, manifestJson, 0600)
	require.NoError(suite.T(), err)

	return dir
}

func (suite *DirSuite) AssertCurrentPointsTo(version string) {
	target, err := os.Readlink(suite.CurrentSymlink())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), target, suite.ReleaseDir(version))
}

// assert current symlink is actually a symlink
func (suite *DirSuite) AssertCurrentIsSymlink() {
	symlink := suite.CurrentSymlink()
	fi, err := os.Lstat(symlink)
	require.NoError(suite.T(), err)
	isSymlink := (fi.Mode() & os.ModeSymlink) == os.ModeSymlink
	assert.True(suite.T(), isSymlink)
}

func (suite *DirSuite) ReleaseDir(version string) string {
	return path.Join(suite.dirPath, version)
}

func (suite *DirSuite) CurrentSymlink() string {
	return path.Join(suite.dirPath, currentSymlink)
}

func (suite *DirSuite) CreateReleaseDir(version string) {
	releaseDir := suite.ReleaseDir(version)
	require.NoError(suite.T(), os.Mkdir(releaseDir, 0755))
}

func (suite *DirSuite) SetCurrentSymlink(version string) {
	target := suite.ReleaseDir(version)
	link := suite.CurrentSymlink()
	require.NoError(suite.T(), os.Symlink(target, link))
}

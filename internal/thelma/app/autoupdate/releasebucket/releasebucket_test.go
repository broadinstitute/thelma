package releasebucket

import (
	"encoding/json"
	"fmt"
	scratchmocks "github.com/broadinstitute/thelma/internal/thelma/app/scratch/mocks"
	bucketmocks "github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/testing/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"os/exec"
	"path"
	"testing"
)

const testArchive = "testdata/thelma_v1.2.3_darwin_arm64.tar.gz"
const testArchiveSha256Sum = "2c4d0ccea229b7388fc0ed7f077c9516fd5590ad488ec6ec45406d73d2f002db"
const testArchiveVersion = "v1.2.3"

type ReleaseBucketSuite struct {
	suite.Suite
	gcsBucket *bucketmocks.Bucket
	runner    *shell.MockRunner
	scratch   *scratchmocks.Scratch
	bucket    ReleaseBucket
}

func TestReleaseBucket(t *testing.T) {
	suite.Run(t, new(ReleaseBucketSuite))
}

func (suite *ReleaseBucketSuite) SetupTest() {
	t := suite.T()

	gcsBucket := bucketmocks.NewBucket(t)

	mockRunner := shell.DefaultMockRunner()

	_scratch := scratchmocks.NewScratch(t)

	_bucket := New(gcsBucket, mockRunner, _scratch)

	suite.gcsBucket = gcsBucket
	suite.runner = mockRunner
	suite.scratch = _scratch
	suite.bucket = _bucket
}

func (suite *ReleaseBucketSuite) TestResolveTagOrVersion() {
	suite.ExpectReadTags(map[string]string{
		"latest": "v1.2.3",
		"v2":     "v2.0.1",
	})

	_bucket := suite.bucket

	resolved, err := _bucket.ResolveTagOrVersion("latest")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v1.2.3", resolved)

	resolved, err = _bucket.ResolveTagOrVersion("v2")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v2.0.1", resolved)

	resolved, err = _bucket.ResolveTagOrVersion("4.5.6")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v4.5.6", resolved)

	_, err = _bucket.ResolveTagOrVersion("missing")
	require.Error(suite.T(), err)
	assert.ErrorContains(suite.T(), err, "not a valid Thelma tag or semantic version")
}

func (suite *ReleaseBucketSuite) TestDownloadAndUnpack() {
	archive := NewArchive(testArchiveVersion)

	tmpdir := suite.T().TempDir()

	suite.scratch.EXPECT().Mkdir(tmpdirName).Return(tmpdir, nil)

	suite.gcsBucket.EXPECT().Exists(archive.ObjectPath()).Return(true, nil)

	suite.gcsBucket.EXPECT().Download(archive.ObjectPath(), path.Join(tmpdir, archive.Filename())).Run(func(objectName string, localPath string) {
		content, err := os.ReadFile(testArchive)
		require.NoError(suite.T(), err)
		err = os.WriteFile(path.Join(tmpdir, path.Base(objectName)), content, 0600)
		require.NoError(suite.T(), err)
	}).Return(nil)

	checksumContent := fmt.Sprintf("%s\t%s", testArchiveSha256Sum, archive.Filename())
	suite.gcsBucket.EXPECT().Read(archive.Sha256SumObjectPath()).Return([]byte(checksumContent), nil)

	suite.runner.ExpectCmd(shell.Command{Prog: "tar", Args: []string{
		"-xz",
		"-C", path.Join(tmpdir, archive.Version()),
		"-f", path.Join(tmpdir, archive.Filename()),
	}}).Run(func(args mock.Arguments) {
		cmd := args.Get(0).(shell.Command)
		// simulating the behavior of tar in go is hard, so... uh... just run the command.
		// we could _not_ mock the runner but this helps us prevent suprise commands from being executed
		// and verify that `tar` is called w/ correct args, so :shrug:
		require.NoError(suite.T(), exec.Command(cmd.Prog, cmd.Args...).Run())
	}).Return(nil)

	unpackDir, err := suite.bucket.DownloadAndUnpack(archive)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), path.Join(tmpdir, testArchiveVersion), unpackDir)

	content, err := os.ReadFile(path.Join(unpackDir, "hello.txt"))
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "hello\n", string(content))
}

func (suite *ReleaseBucketSuite) ExpectReadTags(tagsToReturn map[string]string) {
	content, err := json.Marshal(tagsToReturn)
	require.NoError(suite.T(), err)
	suite.gcsBucket.EXPECT().Read(tagsFile).Return(content, nil)
}

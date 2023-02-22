//go:build smoke
// +build smoke

package spawn

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

// the spawn tester checks this and uses it to figure out where to write logs
const fakeRootEnvVar = "FAKE_THELMA_ROOT"

// the spawn tester checks this and uses it to figure out how to name logs
const fakeLogFileNameEnvVar = "FAKE_LOGFILE_NAME"

const logFileName = "spawn-tests"
const sleepInterval = 50 * time.Millisecond

type SpawnSuite struct {
	suite.Suite
	fakeRoot    string
	spawnTester string
	spawn       Spawn
}

func TestSpawn(t *testing.T) {
	suite.Run(t, new(SpawnSuite))
}

func (suite *SpawnSuite) SetupSuite() {
	// compile the spawn tester in testing/main to <tmpdir>/spawn-tester
	spawnTester := path.Join(suite.T().TempDir(), "spawn-tester")
	err := exec.Command("go", "build", "-o", spawnTester, "./testing/main").Run()
	require.NoError(suite.T(), err)
	suite.spawnTester = spawnTester
}

func (suite *SpawnSuite) SetupTest() {
	suite.fakeRoot = suite.T().TempDir()
}

func (suite *SpawnSuite) Test_Spawn_SucceedsFastCommand() {
	testfile := path.Join(suite.T().TempDir(), "testfile")
	suite.RunSpawnTester("touch", testfile)
	suite.AssertFileExistsWithinDuration(testfile, 5*time.Second)
}

func (suite *SpawnSuite) Test_Spawn_SucceedsSlowCommand() {
	testfile := path.Join(suite.T().TempDir(), "testfile")
	suite.RunSpawnTester("sh", "-c", fmt.Sprintf("sleep 3 && touch %s", testfile))
	suite.AssertFileExistsWithinDuration(testfile, 10*time.Second)
}

func (suite *SpawnSuite) Test_Spawn_LogsOutput() {
	testfile := path.Join(suite.T().TempDir(), "testfile")
	suite.RunSpawnTester("sh", "-c", fmt.Sprintf(`echo "an error" >&2 && echo "some output" && touch %s`, testfile))
	suite.AssertFileExistsWithinDuration(testfile, 5*time.Second)
	assert.Equal(suite.T(), "an error\n", suite.ReadStderrForSpawnedCommand())
	assert.Equal(suite.T(), "some output\n", suite.ReadStdoutForSpawnedCommand())
}

func (suite *SpawnSuite) Test_Spawn_PanicsOnRecursion() {
	suite.RunSpawnTester(suite.spawnTester, "echo", "hello")
	time.Sleep(500 * time.Millisecond)
	stderr := suite.ReadStderrForSpawnedCommand()
	assert.Regexp(suite.T(), "won't spawn child process.*already a Thelma spawn", stderr)
}

func (suite *SpawnSuite) AssertFileExistsWithinDuration(file string, duration time.Duration) {
	stopAt := time.Now().Add(duration)
	for time.Now().Before(stopAt) {
		exists, err := utils.FileExists(file)
		require.NoError(suite.T(), err)
		if exists {
			assert.FileExists(suite.T(), file, "file %s must exist within %s", file, duration)
			return
		}
		time.Sleep(sleepInterval)
	}

	assert.FileExists(suite.T(), file, "file %s must exist within %s", file, duration)
}

func (suite *SpawnSuite) ReadStdoutForSpawnedCommand() string {
	return suite.ReadLogFile(stdoutLogExt)
}

func (suite *SpawnSuite) ReadStderrForSpawnedCommand() string {
	return suite.ReadLogFile(stderrLogExt)
}

func (suite *SpawnSuite) ReadLogFile(ext string) string {
	content, err := os.ReadFile(path.Join(suite.fakeRoot, "logs", logFileName+ext))
	require.NoError(suite.T(), err)
	return string(content)
}

func (suite *SpawnSuite) RunSpawnTester(args ...string) {
	cmd := exec.Command(suite.spawnTester, args...)

	// make sure the spawn tester records logs in our configured fake root
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fakeRootEnvVar+"="+suite.fakeRoot)
	cmd.Env = append(cmd.Env, fakeLogFileNameEnvVar+"="+logFileName)

	log.Info().Msgf("fake root: %s", suite.fakeRoot)
	log.Info().Str("path", suite.spawnTester).Str("args", utils.QuoteJoin(args)).Msgf("Running spawn tester...")
	err := cmd.Run()
	require.NoError(suite.T(), err)
}

// okay so-
// - compile the spawn tester
// - to test:
//   - success for short subprocess
//   - success for longer subprocess
//   - panic on recursion
// - also, should probably tag this as a smoke test
// - make sure it executes the subprocess

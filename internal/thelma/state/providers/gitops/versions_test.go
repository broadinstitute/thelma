package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"os/exec"
	"path"
	"testing"
)

func TestSnapshot_ChartVersion(t *testing.T) {
	thelmaHome := t.TempDir()
	runner := shell.DefaultMockRunner()

	err := initializeFakeVersionsDir(thelmaHome)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	v, err := NewVersions(thelmaHome)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	s1 := v.GetSnapshot(terra.AppReleaseType, Dev)

	assert.True(t, s1.ReleaseDefined("agora"))
	assert.Equal(t, "0.10.0", s1.ChartVersion("agora"))
	assert.Equal(t, "v100", s1.AppVersion("agora"))

	s2 := v.GetSnapshot(terra.ClusterReleaseType, Alpha)
	assert.True(t, s2.ReleaseDefined("prometheus"))
	assert.Equal(t, "0.1.3", s2.ChartVersion("prometheus"))
	assert.Equal(t, "", s2.AppVersion("prometheus"))

	runner.AssertExpectations(t)
}

func TestReleaseType_String(t *testing.T) {
	assert.Equal(t, "app", terra.AppReleaseType.String())
	assert.Equal(t, "cluster", terra.ClusterReleaseType.String())
}

func TestReleaseType_UnmarshalYAML(t *testing.T) {
	var err error
	var r terra.ReleaseType

	err = yaml.Unmarshal([]byte("app"), &r)
	assert.NoError(t, err)
	assert.Equal(t, terra.AppReleaseType, r)

	err = yaml.Unmarshal([]byte("cluster"), &r)
	assert.NoError(t, err)
	assert.Equal(t, terra.ClusterReleaseType, r)

	err = yaml.Unmarshal([]byte("invalid"), &r)
	assert.Error(t, err)
	assert.Regexp(t, "unknown release type", err)
}

func TestMocks_MatchInterface(t *testing.T) {
	v := NewMockVersions()
	s := NewMockSnapshot()

	// make sure interfaces match -- compilation will fail if they don't
	var _ Versions = v
	var _ VersionSnapshot = s
}

func initializeFakeVersionsDir(thelmaHome string) error {
	cmd := exec.Command("cp", "-r", "testdata/versions", path.Join(thelmaHome, "versions"))
	return cmd.Run()
}

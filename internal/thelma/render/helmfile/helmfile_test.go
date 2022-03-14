package helmfile

import (
	"github.com/broadinstitute/thelma/internal/thelma/render/resolver"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testState struct {
	configRepo *ConfigRepo
	mockRunner *shell.MockRunner
}

func TestHelmfileUpdate(t *testing.T) {
	testCases := []struct {
		description string
		setupMocks  func(t *testing.T, ts *testState)
	}{
		{
			description: "info level logging",
			setupMocks: func(t *testing.T, ts *testState) {
				ts.configRepo.helmfileLogLevel = "info"
				ts.mockRunner.ExpectCmd(shell.Command{
					Prog: "helmfile",
					Args: []string{
						"--log-level=info",
						"--allow-no-matching-release",
						"repos",
					},
					Dir: ts.configRepo.thelmaHome,
				})
			},
		},
		{
			description: "debug level logging",
			setupMocks: func(t *testing.T, ts *testState) {
				ts.configRepo.helmfileLogLevel = "debug"
				ts.mockRunner.ExpectCmd(shell.Command{
					Prog: "helmfile",
					Args: []string{
						"--log-level=debug",
						"--allow-no-matching-release",
						"repos",
					},
					Dir: ts.configRepo.thelmaHome,
				})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			ts := setupTestState(t)
			testCase.setupMocks(t, ts)

			err := ts.configRepo.HelmUpdate()
			assert.NoError(t, err)

			ts.mockRunner.AssertExpectations(t)
		})
	}
}

func setupTestState(t *testing.T) *testState {
	mockRunner := shell.DefaultMockRunner()
	mockRunner.Test(t)

	configRepo := NewConfigRepo(Options{
		ThelmaHome:       t.TempDir(),
		ChartCacheDir:    t.TempDir(),
		HelmfileLogLevel: "info",
		ShellRunner:      mockRunner,
		ResolverMode:     resolver.Deploy,
	})

	return &testState{
		mockRunner: mockRunner,
		configRepo: configRepo,
	}
}

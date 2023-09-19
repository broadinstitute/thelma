package filetrigger

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// ChartList is a good entrypoint for testing since it exercises all the other functions in this package
func Test_ChartList(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		expected   []string
		errMatcher string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "sam files only",
			input:    "charts/sam/templates/sam-deployment.yaml",
			expected: []string{"sam"},
		},
		{
			name: "sam and rawls files",
			input: `
charts/sam/templates/deployment.yaml
charts/rawls/values.yaml
`,
			expected: []string{"rawls", "sam"},
		},
		{
			name: "values files",
			input: `
values/app/rawls.yaml.gotmpl
values/app/sam/live.yaml.gotmpl
values/cluster/yale/terra.yaml
`,
			expected: []string{"rawls", "sam", "yale"},
		},
		{
			name: "global app release values - root file",
			input: `
values/app/global.yaml.gotmpl
`,
			expected: []string{"rawls", "sam", "workspacemanager"},
		},
		{
			name: "global app release values - env override file",
			input: `
values/app/global/live.yaml.gotmpl
`,
			expected: []string{"rawls", "sam", "workspacemanager"},
		},
		{
			name: "all cluster releases",
			input: `
values/cluster/global/terra.yaml
`,
			expected: []string{"secrets-manager", "yale"},
		},
		{
			name: "helmfile.yaml",
			input: `
helmfile.yaml
`,
			expected: []string{"rawls", "sam", "secrets-manager", "workspacemanager", "yale"},
		},
		{
			name: "app and cluster globals",
			input: `
values/app/global/live/dev.yaml
values/cluster/global.yaml.gotmpl
`,
			expected: []string{"rawls", "sam", "secrets-manager", "workspacemanager", "yale"},
		},
		{
			name: "cluster global values and rawls value",
			input: `
values/app/rawls.yaml.gotmpl
values/cluster/global.yaml.gotmpl
`,
			expected: []string{"rawls", "secrets-manager", "yale"},
		},
		{
			name: "app global values and yale value",
			input: `
values/app/yale/bee.yaml.gotmpl
values/app/global/bee.yaml.gotmpl
`,
			expected: []string{"rawls", "sam", "workspacemanager", "yale"},
		},
		{
			name: "error on non-relative path",
			input: `
/Users/jdoe/terra-helmfile/helmfile.yaml
`,
			errMatcher: "contains absolute path",
		},
		{
			name: "chart with no corresponding release",
			input: `
charts/ingress/some/file.txt
charts/rawls/values.yaml
`,
			expected: []string{"ingress", "rawls"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture, err := statefixtures.LoadFixtureFromFile("testdata/statefixture.yaml")
			require.NoError(t, err)

			inputFile := t.TempDir() + "/input.txt"
			require.NoError(t, os.WriteFile(inputFile, []byte(tc.input), 0600))

			actual, err := ChartList(inputFile, fixture.Mocks().State)

			if tc.errMatcher != "" {
				require.Error(t, err, "expected error matching %s", tc.errMatcher)
				assert.ErrorContains(t, err, tc.errMatcher)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

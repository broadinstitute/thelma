package changedfiles

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sort"
	"testing"
)

func Test_ChangedList(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectCharts   []string
		expectReleases []string
		errMatcher     string
	}{
		{
			name:           "empty input",
			input:          "",
			expectCharts:   nil,
			expectReleases: nil,
		},
		{
			name:           "sam files only",
			input:          "charts/sam/templates/sam-deployment.yaml",
			expectCharts:   []string{"sam"},
			expectReleases: []string{"sam-dev"},
		},
		{
			name: "sam and rawls files",
			input: `
charts/sam/templates/deployment.yaml
charts/rawls/values.yaml
`,
			expectCharts:   []string{"rawls", "sam"},
			expectReleases: []string{"rawls-staging", "sam-dev"},
		},
		{
			name: "values files",
			input: `
values/app/rawls.yaml.gotmpl
values/app/sam/live.yaml.gotmpl
values/cluster/yale/terra.yaml
`,
			expectCharts:   []string{"rawls", "sam", "yale"},
			expectReleases: []string{"rawls-staging", "sam-dev", "yale-terra-dev", "yale-terra-staging"},
		},
		{
			name: "global app release values - root file",
			input: `
values/app/global.yaml.gotmpl
`,
			expectCharts:   []string{"rawls", "sam", "workspacemanager"},
			expectReleases: []string{"datarepo-my-bee", "rawls-staging", "sam-dev", "workspacemanager-swatomation"},
		},
		{
			name: "global app release values - env override file",
			input: `
values/app/global/live.yaml.gotmpl
`,
			expectCharts:   []string{"rawls", "sam", "workspacemanager"},
			expectReleases: []string{"datarepo-my-bee", "rawls-staging", "sam-dev", "workspacemanager-swatomation"},
		},
		{
			name: "all cluster releases",
			input: `
values/cluster/global/terra.yaml
`,
			expectCharts:   []string{"secrets-manager", "yale"},
			expectReleases: []string{"secrets-manager-terra-dev", "yale-terra-dev", "yale-terra-staging"},
		},
		{
			name: "helmfile.yaml",
			input: `
helmfile.yaml
`,
			expectCharts:   []string{"rawls", "sam", "secrets-manager", "workspacemanager", "yale"},
			expectReleases: []string{"datarepo-my-bee", "rawls-staging", "sam-dev", "secrets-manager-terra-dev", "workspacemanager-swatomation", "yale-terra-dev", "yale-terra-staging"},
		},
		{
			name: "app and cluster globals",
			input: `
values/app/global/live/dev.yaml
values/cluster/global.yaml.gotmpl
`,
			expectCharts:   []string{"rawls", "sam", "secrets-manager", "workspacemanager", "yale"},
			expectReleases: []string{"datarepo-my-bee", "rawls-staging", "sam-dev", "secrets-manager-terra-dev", "workspacemanager-swatomation", "yale-terra-dev", "yale-terra-staging"},
		},
		{
			name: "cluster global values and rawls value",
			input: `
values/app/rawls.yaml.gotmpl
values/cluster/global.yaml.gotmpl
`,
			expectCharts:   []string{"rawls", "secrets-manager", "yale"},
			expectReleases: []string{"rawls-staging", "secrets-manager-terra-dev", "yale-terra-dev", "yale-terra-staging"},
		},
		{
			name: "app global values and yale value",
			input: `
values/cluster/yale/bee.yaml.gotmpl
values/app/global/bee.yaml.gotmpl
`,
			expectCharts:   []string{"rawls", "sam", "workspacemanager", "yale"},
			expectReleases: []string{"datarepo-my-bee", "rawls-staging", "sam-dev", "workspacemanager-swatomation", "yale-terra-dev", "yale-terra-staging"},
		},
		{
			name: "error on non-relative path",
			input: `
/Users/jdoe/terra-helmfile/helmfile.yaml
`,
			errMatcher: "contains absolute path",
		},
		{
			name: "chart with downstream dependents - ingress",
			input: `
charts/ingress/some/file.txt
`,
			expectCharts:   []string{"agora", "foundation", "ingress", "rawls", "sam", "workspacemanager"},
			expectReleases: []string{"rawls-staging", "sam-dev", "workspacemanager-swatomation"},
		},
		{
			name: "chart with downstream dependents - postgres",
			input: `
charts/postgres/some/file.txt
`,
			expectCharts:   []string{"foundation", "postgres", "sam", "workspacemanager"},
			expectReleases: []string{"sam-dev", "workspacemanager-swatomation"},
		},
		{
			name: "chart with downstream dependents - mysql",
			input: `
charts/mysql/some/file.txt
`,
			expectCharts:   []string{"agora", "mysql", "rawls"},
			expectReleases: []string{"rawls-staging"},
		},
		{
			name: "chart with downstream dependents - foundation",
			input: `
charts/foundation/some/file.txt
`,
			expectCharts:   []string{"foundation", "workspacemanager"},
			expectReleases: []string{"workspacemanager-swatomation"},
		},
		{
			name: "release with chart that lives outside terra-helmfile - i.e. datarepo",
			input: `
values/app/datarepo.yaml
`,
			expectCharts:   nil,
			expectReleases: []string{"datarepo-my-bee"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRunner := shell.DefaultMockRunner()

			fixture, err := statefixtures.LoadFixtureFromFile("testdata/statefixture.yaml")
			require.NoError(t, err)

			chartsDir, err := source.NewChartsDir("testdata/charts", mockRunner)
			require.NoError(t, err)

			changed := New(chartsDir, fixture.Mocks().State)

			inputFile := t.TempDir() + "/input.txt"
			require.NoError(t, os.WriteFile(inputFile, []byte(tc.input), 0600))

			actualCharts, err := changed.ChartList(inputFile)

			if tc.errMatcher != "" {
				require.Error(t, err, "expected error matching %s", tc.errMatcher)
				assert.ErrorContains(t, err, tc.errMatcher)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectCharts, actualCharts)
			}

			actualReleaseFilter, err := changed.ReleaseFilter(inputFile)
			if tc.errMatcher != "" {
				require.Error(t, err, "expected error matching %s", tc.errMatcher)
				assert.ErrorContains(t, err, tc.errMatcher)
			} else {
				require.NoError(t, err)
				matches, err := fixture.Mocks().Releases.Filter(actualReleaseFilter)
				require.NoError(t, err)
				var releaseNames []string
				for _, r := range matches {
					releaseNames = append(releaseNames, r.FullName())
				}
				sort.Strings(releaseNames)
				assert.Equal(t, tc.expectReleases, releaseNames)
			}

		})
	}
}

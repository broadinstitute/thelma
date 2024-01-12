package source

import (
	toolboxtesting "github.com/broadinstitute/thelma/internal/thelma/toolbox/testing"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os/exec"
	"path"
	"sort"
	"testing"
)

type ChartsDirTestSuite struct {
	suite.Suite
	dir       string
	runner    *shell.MockRunner
	chartsDir ChartsDir
}

func (suite *ChartsDirTestSuite) SetupTest() {
	suite.dir = path.Join(suite.T().TempDir(), "charts")
	err := exec.Command("cp", "-r", "testdata/charts", suite.dir).Run()
	require.NoError(suite.T(), err)

	suite.runner = shell.NewMockRunner(shell.MockOptions{
		VerifyOrder: false,
	})
	chartsDir, err := NewChartsDir(suite.dir, suite.runner)
	require.NoError(suite.T(), err)
	suite.chartsDir = chartsDir
}

func (suite *ChartsDirTestSuite) TestPath() {
	assert.Equal(suite.T(), suite.dir, suite.chartsDir.Path())
}

func (suite *ChartsDirTestSuite) TestExists() {
	assert.True(suite.T(), suite.chartsDir.Exists("agora"))
	assert.True(suite.T(), suite.chartsDir.Exists("foundation"))
	assert.True(suite.T(), suite.chartsDir.Exists("yale"))
	assert.False(suite.T(), suite.chartsDir.Exists("missing"))
}

func (suite *ChartsDirTestSuite) TestGetChart() {
	chart, err := suite.chartsDir.GetChart("agora")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "agora", chart.Name())
	assert.Equal(suite.T(), "1.2.3", chart.ManifestVersion())

	_, err = suite.chartsDir.GetChart("missing")
	assert.Error(suite.T(), err)
	assert.ErrorContains(suite.T(), err, `chart "missing" does not exist in source dir`)
}

func (suite *ChartsDirTestSuite) TestGetCharts() {
	charts, err := suite.chartsDir.GetCharts("agora", "foundation", "ingress", "yale")
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), charts, 4)
	assert.Equal(suite.T(), "agora", charts[0].Name())
	assert.Equal(suite.T(), "1.2.3", charts[0].ManifestVersion())
	assert.Equal(suite.T(), "foundation", charts[1].Name())
	assert.Equal(suite.T(), "1.2.3", charts[1].ManifestVersion())
	assert.Equal(suite.T(), "ingress", charts[2].Name())
	assert.Equal(suite.T(), "1.2.3", charts[2].ManifestVersion())
	assert.Equal(suite.T(), "yale", charts[3].Name())
	assert.Equal(suite.T(), "1.2.3", charts[3].ManifestVersion())

	_, err = suite.chartsDir.GetCharts("agora", "foundation", "ingress", "missing", "yale")
	assert.Error(suite.T(), err)
	assert.ErrorContains(suite.T(), err, `chart "missing" does not exist in source dir`)
}

func (suite *ChartsDirTestSuite) TestUpdateDependentVersionConstraints() {
	finder, err := toolboxtesting.NewToolFinderForTests()
	require.NoError(suite.T(), err)
	realRunner := shell.NewRunner(finder)

	// agora has no dependents, so no yq commands will be run
	chart, err := suite.chartsDir.GetChart("agora")
	require.NoError(suite.T(), err)
	err = suite.chartsDir.UpdateDependentVersionConstraints(chart, "4.5.6")
	require.NoError(suite.T(), err)

	chart, err = suite.chartsDir.GetChart("ingress")
	require.NoError(suite.T(), err)

	// ingress chart multiple dependents; all should be updated
	ingressDependents := []string{
		"foundation",
		"agora",
		"rawls",
		"sam",
	}

	for _, dependent := range ingressDependents {
		cmd := shell.Command{
			Prog: "yq",
			Args: []string{
				"eval",
				"--inplace", "(.dependencies.[] | select(.name == \"ingress\") | .version) |= \"4.5.6\"",
				suite.dir + "/" + dependent + "/Chart.yaml",
			},
		}
		suite.runner.ExpectCmd(cmd).Run(func(_ mock.Arguments) {
			// this is hacky, but the Chart implementation is strict; it actually checks for
			// side effects after the command is run.
			// and the easiest way to create those side effects is to just run the command.
			assert.NoError(suite.T(), realRunner.Run(cmd))
		}).Return(nil)
	}

	err = suite.chartsDir.UpdateDependentVersionConstraints(chart, "4.5.6")
	require.NoError(suite.T(), err)
}

func (suite *ChartsDirTestSuite) TestWithTransitiveDependents() {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no dependents",
			input:    []string{"agora"},
			expected: []string{"agora"},
		},
		{
			name:     "no dependents multiple",
			input:    []string{"agora", "rawls", "sam", "yale"},
			expected: []string{"agora", "rawls", "sam", "yale"},
		},
		{
			name:     "1 layer - foundation",
			input:    []string{"foundation"},
			expected: []string{"foundation", "workspacemanager"},
		},
		{
			name:     "1 layer - mysql",
			input:    []string{"mysql"},
			expected: []string{"agora", "mysql", "rawls"},
		},
		{
			name:     "2 layers - ingress",
			input:    []string{"ingress"},
			expected: []string{"agora", "foundation", "ingress", "rawls", "sam", "workspacemanager"},
		},
		{
			name:     "2 layers - postgres",
			input:    []string{"postgres"},
			expected: []string{"foundation", "postgres", "sam", "workspacemanager"},
		},
		{
			name:     "combined - postgres, mysql",
			input:    []string{"mysql", "postgres"},
			expected: []string{"agora", "foundation", "mysql", "postgres", "rawls", "sam", "workspacemanager"},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			charts, err := suite.chartsDir.GetCharts(tc.input...)
			require.NoError(t, err)
			withDeps, err := suite.chartsDir.WithTransitiveDependents(charts)
			require.NoError(t, err)
			names := ChartNames(withDeps...)
			sort.Strings(names)

			assert.Equal(t, tc.expected, names)
		})
	}
}

func (suite *ChartsDirTestSuite) TestRecursivelyUpdateDependencies() {
	yale, err := suite.chartsDir.GetChart("yale")
	require.NoError(suite.T(), err)

	agora, err := suite.chartsDir.GetChart("agora")
	require.NoError(suite.T(), err)

	rawls, err := suite.chartsDir.GetChart("rawls")
	require.NoError(suite.T(), err)

	workspacemanager, err := suite.chartsDir.GetChart("workspacemanager")
	require.NoError(suite.T(), err)

	suite.expectDependencyUpdate("yale")
	require.NoError(suite.T(), suite.chartsDir.RecursivelyUpdateDependencies(yale))

	suite.expectDependencyUpdate("agora", "mysql", "ingress")
	require.NoError(suite.T(), suite.chartsDir.RecursivelyUpdateDependencies(agora))

	suite.expectDependencyUpdate("workspacemanager", "foundation", "ingress", "postgres")
	require.NoError(suite.T(), suite.chartsDir.RecursivelyUpdateDependencies(workspacemanager))

	suite.expectDependencyUpdate("agora", "mysql", "ingress", "rawls", "yale")
	require.NoError(suite.T(), suite.chartsDir.RecursivelyUpdateDependencies(agora, rawls, yale))
}

func (suite *ChartsDirTestSuite) expectDependencyUpdate(chartNames ...string) {
	for _, chartName := range chartNames {
		cmd := shell.Command{
			Prog: "helm",
			Args: []string{
				"dependency",
				"update",
				"--skip-refresh",
			},
			Dir: path.Join(suite.dir, chartName),
		}
		suite.runner.ExpectCmd(cmd).Return(nil)
	}
}

func TestChartsDirSuite(t *testing.T) {
	suite.Run(t, new(ChartsDirTestSuite))
}

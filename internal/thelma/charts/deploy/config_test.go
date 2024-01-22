package deploy

import (
	sourcemocks "github.com/broadinstitute/thelma/internal/thelma/charts/source/mocks"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/broadinstitute/thelma/internal/thelma/utils/stateutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type ConfigSuite struct {
	suite.Suite
	chartsDir    *sourcemocks.ChartsDir
	state        *statemocks.State
	configLoader ConfigLoader
}

func (suite *ConfigSuite) SetupSubTest() {
	statefixture, err := statefixtures.LoadFixtureFromFile("testdata/statefixture.yaml")
	require.NoError(suite.T(), err)

	suite.state = statefixture.Mocks().State

	suite.chartsDir = sourcemocks.NewChartsDir(suite.T())

	loader, err := newConfigLoader(suite.chartsDir, suite.state)
	require.NoError(suite.T(), err)
	suite.configLoader = loader
}

func (suite *ConfigSuite) Test_findReleases() {
	testCases := []struct {
		name           string
		chartName      string
		expectReleases []string
		configFile     string
	}{
		{
			name:           "chart has no releases in state",
			chartName:      "not-in-state",
			expectReleases: []string{},
		},
		{
			name:           "chart with no .autorelease.yaml should default to dev",
			chartName:      "agora",
			expectReleases: []string{"agora-dev"},
		},
		{
			name:           "chart with invalid .autorelease.yaml should fallback to default",
			chartName:      "agora",
			configFile:     `"unterminated string`,
			expectReleases: []string{"agora-dev"},
		},
		{
			name:      "chart with release in .autorelease.yaml that does not exist",
			chartName: "agora",
			configFile: `
sherlock:
  chartReleasesToUseLatest:
    - agora-doesnotexist
`,
			expectReleases: []string{},
		},
		{
			name:      "chart should use releases in .autorelease.yaml if specified",
			chartName: "agora",
			configFile: `
sherlock:
  chartReleasesToUseLatest:
    - agora-staging
`,
			expectReleases: []string{"agora-staging"},
		},
		{
			name:      "chart should use all releases in .autorelease.yaml if specified",
			chartName: "agora",
			configFile: `
sherlock:
  chartReleasesToUseLatest:
    - agora-dev
    - agora-staging
`,
			expectReleases: []string{"agora-dev", "agora-staging"},
		},
		{
			name:      "chart should have no releases to update if .autorelease.yaml has enabled: false",
			chartName: "agora",
			configFile: `
enabled: false
`,
			expectReleases: []string{},
		},
		{
			name:           "chart should have no releases if no dev release",
			chartName:      "yale",
			expectReleases: []string{},
		},
		{
			name:      "cluster releases should include all specified releases that exist",
			chartName: "yale",
			configFile: `
sherlock:
  chartReleasesToUseLatest:
    - yale-terra-qa-bees
    - yale-terra-staging
`,
			expectReleases: []string{"yale-terra-qa-bees"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			chartDir := suite.T().TempDir()
			chart := sourcemocks.NewChart(suite.T())
			chart.EXPECT().Path().Return(chartDir)
			chart.EXPECT().Name().Return(tc.chartName).Maybe()
			suite.chartsDir.EXPECT().GetChart(tc.chartName).Return(chart, nil)

			if len(tc.configFile) > 0 {
				err := os.WriteFile(chartDir+"/"+configFile, []byte(tc.configFile), 0644)
				require.NoError(suite.T(), err)
			}

			releases, err := suite.configLoader.FindReleasesToUpdate(tc.chartName)
			require.NoError(suite.T(), err)
			assert.ElementsMatch(suite.T(), tc.expectReleases, stateutils.ReleaseFullNames(releases))
		})
	}
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

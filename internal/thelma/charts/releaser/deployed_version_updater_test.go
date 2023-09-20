package releaser

import (
	"fmt"
	sourcemocks "github.com/broadinstitute/thelma/internal/thelma/charts/source/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	sherlockmocks "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock/mocks"

	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestAutoReleaser_UpdateVersionsFile(t *testing.T) {
	chartName := "mychart"
	newVersion := "5.6.7"
	lastVersion := "5.6.6"
	description := "my new description"

	type mocks struct {
		sherlockUpdater *sherlockmocks.ChartVersionUpdater
	}

	testCases := []struct {
		name          string
		newVersion    string
		configContent string
		setupMocks    func(mocks)
		matchErr      string
	}{
		{
			name: "No config file should default to enabled + app release type",
			setupMocks: func(m mocks) {
				m.sherlockUpdater.On("UpdateForNewChartVersion", chartName, newVersion, lastVersion, description,
					fmt.Sprintf("%s/%s", targetEnvironment, chartName)).Return(nil)
			},
		},
		{
			name:          "Should not update release version if disabled in config file",
			configContent: `enabled: false`,
		},
		{
			name:          "Should support release name overriding",
			configContent: `release: {name: foo}`,
			setupMocks: func(m mocks) {
				m.sherlockUpdater.On("UpdateForNewChartVersion", "foo", newVersion, lastVersion, description,
					fmt.Sprintf("%s/%s", targetEnvironment, "foo")).Return(nil)
			},
		},
		{
			name:          "Should support release type overriding",
			configContent: `release: {type: cluster}`,
			setupMocks: func(m mocks) {
				m.sherlockUpdater.On("UpdateForNewChartVersion", chartName, newVersion, lastVersion, description,
					fmt.Sprintf("%s/%s", targetEnvironment, chartName)).Return(nil)
			},
		},
		{
			name: "Should support new Sherlock configuration",
			configContent: `
release:
  name: foo
sherlock:
  chartReleasesToUseLatest:
    - dev/bar
    - terra-dev/default/baz
`,
			setupMocks: func(m mocks) {
				m.sherlockUpdater.On("UpdateForNewChartVersion", "foo", newVersion, lastVersion, description,
					"dev/bar", "terra-dev/default/baz").Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		chartDir := t.TempDir()
		chart := sourcemocks.NewChart(t)
		chart.EXPECT().Name().Return("mychart")
		chart.EXPECT().Path().Return(chartDir)

		t.Run(tc.name, func(t *testing.T) {
			m := mocks{
				sherlockUpdater: sherlockmocks.NewChartVersionUpdater(t),
			}
			if tc.setupMocks != nil {
				tc.setupMocks(m)
			}

			if len(tc.configContent) > 0 {
				if err := os.WriteFile(path.Join(chartDir, configFile), []byte(tc.configContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			updater := &DeployedVersionUpdater{SherlockUpdaters: []sherlock.ChartVersionUpdater{m.sherlockUpdater}}
			// lastVersion and description are arguments handled solely on Sherlock's end, Thelma doesn't need to even
			// validate them
			err := updater.UpdateReleaseVersion(chart, newVersion, lastVersion, description)

			m.sherlockUpdater.AssertExpectations(t)

			if len(tc.matchErr) == 0 {
				assert.NoError(t, err)
				return
			}

			assert.Error(t, err)
			assert.Regexp(t, tc.matchErr, err)
		})
	}
}

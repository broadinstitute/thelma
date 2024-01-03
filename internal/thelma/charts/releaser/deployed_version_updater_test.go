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
		name               string
		newVersion         string
		configContent      string
		setupMocks         func(mocks)
		expectReleaseNames []string
		matchErr           string
	}{
		{
			name: "No config file should default to enabled + app release type",
			setupMocks: func(m mocks) {
				m.sherlockUpdater.On("UpdateForNewChartVersion", chartName, newVersion, lastVersion, description,
					fmt.Sprintf("%s-%s", chartName, targetEnvironment)).Return(nil)
			},
			expectReleaseNames: []string{fmt.Sprintf("%s-%s", chartName, targetEnvironment)},
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
					fmt.Sprintf("%s-%s", "foo", targetEnvironment)).Return(nil)
			},
			expectReleaseNames: []string{fmt.Sprintf("%s-%s", "foo", targetEnvironment)},
		},
		{
			name:          "Should support release type overriding",
			configContent: `release: {type: cluster}`,
			setupMocks: func(m mocks) {
				m.sherlockUpdater.On("UpdateForNewChartVersion", chartName, newVersion, lastVersion, description,
					fmt.Sprintf("%s-%s", chartName, targetEnvironment)).Return(nil)
			},
			expectReleaseNames: []string{fmt.Sprintf("%s-%s", chartName, targetEnvironment)},
		},
		{
			name: "Should support new Sherlock configuration",
			configContent: `
release:
  name: foo
sherlock:
  chartReleasesToUseLatest:
    - bar-dev
    - baz-terra-dev
`,
			setupMocks: func(m mocks) {
				m.sherlockUpdater.On("UpdateForNewChartVersion", "foo", newVersion, lastVersion, description,
					"bar-dev", "baz-terra-dev").Return(nil)
			},
			expectReleaseNames: []string{"bar-dev", "baz-terra-dev"},
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
			updatedReleaseNames, err := updater.UpdateReleaseVersion(chart, newVersion, lastVersion, description)

			m.sherlockUpdater.AssertExpectations(t)

			if len(tc.matchErr) > 0 {
				assert.Error(t, err)
				assert.Regexp(t, tc.matchErr, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectReleaseNames, updatedReleaseNames)
		})
	}
}

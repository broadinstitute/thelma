package deploy

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/releaser"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	sherlockmocks "github.com/broadinstitute/thelma/internal/thelma/clients/sherlock/mocks"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAutoReleaser_UpdateVersionsFile(t *testing.T) {
	chartName := "mychart"
	newVersion := "5.6.7"
	lastVersion := "5.6.6"
	description := "my new description"
	releaseNames := []string{"mychart-e1", "mychart-dev", "mychart-e2"}

	newUpdater := func(t *testing.T, expectCall bool, err error) *sherlockmocks.ChartVersionUpdater {
		u := sherlockmocks.NewChartVersionUpdater(t)
		if expectCall {
			u.EXPECT().UpdateForNewChartVersion(chartName, newVersion, lastVersion, description, releaseNames).Return(err)
		}
		return u
	}

	testCases := []struct {
		name       string
		setupMocks func(t *testing.T, updater *DeployedVersionUpdater)
		matchErr   string
	}{
		{
			name: "No error if no updaters are configured",
		},
		{
			name: "Call update with all configured updaters",
			setupMocks: func(t *testing.T, updater *DeployedVersionUpdater) {
				updater.SherlockUpdaters = []sherlock.ChartVersionUpdater{
					newUpdater(t, true, nil),
					newUpdater(t, true, nil),
				}
				updater.SoftFailSherlockUpdaters = []sherlock.ChartVersionUpdater{
					newUpdater(t, true, nil),
					newUpdater(t, true, nil),
				}
			},
		},
		{
			name: "Soft-fail update errors should not propagate",
			setupMocks: func(t *testing.T, updater *DeployedVersionUpdater) {
				updater.SherlockUpdaters = []sherlock.ChartVersionUpdater{
					newUpdater(t, true, nil),
					newUpdater(t, true, nil),
				}
				updater.SoftFailSherlockUpdaters = []sherlock.ChartVersionUpdater{
					newUpdater(t, true, nil),
					newUpdater(t, true, errors.Errorf("this should be ignored")),
					newUpdater(t, true, nil),
				}
			},
		},
		{
			name: "Updater errors should propagate",
			setupMocks: func(t *testing.T, updater *DeployedVersionUpdater) {
				updater.SherlockUpdaters = []sherlock.ChartVersionUpdater{
					newUpdater(t, true, nil),
					newUpdater(t, true, errors.Errorf("whoops")),
					newUpdater(t, false, nil),
				}
				updater.SoftFailSherlockUpdaters = []sherlock.ChartVersionUpdater{
					newUpdater(t, false, nil),
					newUpdater(t, false, nil),
				}
			},
			matchErr: "whoops",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			releases := mockReleases(releaseNames)

			updater := &DeployedVersionUpdater{}

			if tc.setupMocks != nil {
				tc.setupMocks(t, updater)
			}

			err := updater.UpdateChartReleaseVersions(
				chartName,
				releases,
				releaser.VersionPair{
					PriorVersion: lastVersion,
					NewVersion:   newVersion,
				},
				description,
			)

			if len(tc.matchErr) > 0 {
				assert.Error(t, err)
				assert.Regexp(t, tc.matchErr, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

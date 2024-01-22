package releaser

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/stateutils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// DeployedVersionUpdater offers a UpdateChartReleaseVersions to take a newly published chart and update development instances to use
// it.
// It stores lists of different update mechanisms so they can be easily enabled/disabled by the caller.
// This is a literal struct, not an interface, so the callers can configure it out without needing to pass multiple
// parameters around.
type DeployedVersionUpdater struct {
	SherlockUpdaters         []sherlock.ChartVersionUpdater
	SoftFailSherlockUpdaters []sherlock.ChartVersionUpdater
}

func (a *DeployedVersionUpdater) ReportNewChartVersion(chartName string, versions VersionPair, description string) error {
	return a.repeatForAllSherlockUpdaters(func(sherlockUpdater sherlock.ChartVersionUpdater) error {
		return sherlockUpdater.ReportNewChartVersion(chartName, versions.NewVersion, versions.PriorVersion, description)
	})
}

func (a *DeployedVersionUpdater) UpdateChartReleaseVersions(chartName string, releases []terra.Release, versions VersionPair, description string) error {
	chartReleaseSelectors := stateutils.ReleaseFullNames(releases)

	return a.repeatForAllSherlockUpdaters(func(sherlockUpdater sherlock.ChartVersionUpdater) error {
		return sherlockUpdater.
			UpdateForNewChartVersion(chartName, versions.NewVersion, versions.PriorVersion, description, chartReleaseSelectors)
	})
}

func (a *DeployedVersionUpdater) repeatForAllSherlockUpdaters(fn func(sherlock.ChartVersionUpdater) error) error {
	for index, sherlockUpdater := range a.SherlockUpdaters {
		err := fn(sherlockUpdater)
		if err != nil {
			return errors.Errorf("autorelease error on sherlock updater %d: %v", index, err)
		}
	}
	for index, sherlockUpdater := range a.SoftFailSherlockUpdaters {
		err := fn(sherlockUpdater)
		if err != nil {
			log.Debug().Err(err).Msgf("autorelease error on sherlock soft-fail updater %d: %v", index, err)
		}
	}
	return nil
}

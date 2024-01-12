package deploy

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/releaser"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
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

func (a *DeployedVersionUpdater) UpdateChartReleaseVersions(chartName string, releases []terra.Release, versions releaser.VersionPair, description string) error {
	chartReleaseSelectors := releaseFullNames(releases)

	for index, sherlockUpdater := range a.SherlockUpdaters {
		err := sherlockUpdater.
			UpdateForNewChartVersion(chartName, versions.NewVersion, versions.PriorVersion, description, chartReleaseSelectors)
		if err != nil {
			return errors.Errorf("autorelease error on sherlock updater %d: %v", index, err)
		}
	}
	for index, sherlockUpdater := range a.SoftFailSherlockUpdaters {
		err := sherlockUpdater.
			UpdateForNewChartVersion(chartName, versions.NewVersion, versions.PriorVersion, description, chartReleaseSelectors)
		if err != nil {
			log.Debug().Err(err).Msgf("autorelease error on sherlock soft-fail updater %d: %v", index, err)
		}
	}

	return nil
}

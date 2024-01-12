package deploy

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/releaser"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sync"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const maxParallelSync = 30

type Options struct {
	DryRun            bool // DryRun if true, don't update sherlock or sync any ArgoCD apps
	IgnoreSyncFailure bool // IgnoreSyncFailure if true, warn about sync failures instead of returning an error
}

type Deployer interface {
	Deploy(chartVersionsToDeploy map[string]releaser.VersionPair, changeDescription string) error
}

func New(chartsDir source.ChartsDir, updater DeployedVersionUpdater, stateLoader terra.StateLoader, syncFactory func() (sync.Sync, error), opts Options) (Deployer, error) {
	state, err := stateLoader.Load()
	if err != nil {
		return nil, err
	}

	cfgLoader, err := newConfigLoader(chartsDir, state)
	if err != nil {
		return nil, err
	}

	return newForTesting(cfgLoader, updater, stateLoader, syncFactory, opts), nil
}

// package-private constructor for testing
func newForTesting(cfgLoader ConfigLoader, updater DeployedVersionUpdater, stateLoader terra.StateLoader, syncFactory func() (sync.Sync, error), opts Options) Deployer {
	return &deployer{
		options:      opts,
		updater:      updater,
		stateLoader:  stateLoader,
		syncFactory:  lazy.NewLazyE[sync.Sync](syncFactory),
		configLoader: cfgLoader,
	}
}

type deployer struct {
	options      Options
	updater      DeployedVersionUpdater
	stateLoader  terra.StateLoader
	syncFactory  lazy.LazyE[sync.Sync]
	configLoader ConfigLoader
}

func (d *deployer) Deploy(chartVersionsToDeploy map[string]releaser.VersionPair, changeDescription string) error {
	syncTargets, err := d.updateSherlock(chartVersionsToDeploy, changeDescription)
	if err != nil {
		return err
	}

	syncTargets, err = d.reloadChartReleases(syncTargets)
	if err != nil {
		return err
	}

	return d.syncArgo(syncTargets)
}

func (d *deployer) updateSherlock(chartVersionsToDeploy map[string]releaser.VersionPair, changeDescription string) ([]terra.Release, error) {
	var syncTargets []terra.Release

	for chartName, versions := range chartVersionsToDeploy {
		releases, err := d.configLoader.FindReleasesToUpdate(chartName)
		if err != nil {
			return nil, errors.Errorf("error identifying releases to update for chart %s: %v", chartName, err)
		}

		if len(releases) == 0 {
			log.Info().Msgf("No releases found in Sherlock for chart %s, skipping", chartName)
			continue
		}

		syncTargets = append(syncTargets, releases...)

		log.Info().Msgf("Updating %d releases in Sherlock for chart %s to version %s: %s", len(releases), chartName, versions.NewVersion, releaseFullNames(releases))
		if d.options.DryRun {
			log.Info().Msg("(skipping update since this is a dry run)")
			continue
		}

		if err = d.updater.UpdateChartReleaseVersions(chartName, releases, versions, changeDescription); err != nil {
			return nil, errors.Errorf("error updating chart releases for %s: %v", chartName, err)
		}
	}

	return syncTargets, nil
}

func (d *deployer) reloadChartReleases(chartReleases []terra.Release) ([]terra.Release, error) {
	state, err := d.stateLoader.Reload()
	if err != nil {
		return nil, err
	}

	allReleases, err := state.Releases().All()
	if err != nil {
		return nil, err
	}

	m := buildReleaseMap(allReleases)

	var reloadedReleases []terra.Release
	for _, r := range chartReleases {
		reloaded, exists := m[r.FullName()]
		if !exists {
			log.Warn().Msgf("updated release %s not found in state, skipping sync", r.FullName())
		}
		reloadedReleases = append(reloadedReleases, reloaded)
	}

	return reloadedReleases, nil
}

func (d *deployer) syncArgo(syncTargets []terra.Release) error {
	log.Info().Msgf("Syncing %d releases...", len(syncTargets))

	if d.options.DryRun {
		log.Info().Msgf("(skipping sync since this is a dry run)")
		return nil
	}

	syncer, err := d.syncFactory.Get()
	if err != nil {
		return errors.Errorf("error creating sync wrapper: %v", err)
	}

	if _, err = syncer.Sync(syncTargets, maxParallelSync); err != nil {
		if d.options.IgnoreSyncFailure {
			log.Warn().Msgf("Error syncing releases: %v", err)
			return nil
		}
		return errors.Errorf("error syncing releases: %v", err)
	}

	log.Info().Msgf("Synced %d releases", len(syncTargets))

	return nil
}

// return the full names of a slice of releases
func releaseFullNames(releases []terra.Release) []string {
	var names []string
	for _, r := range releases {
		names = append(names, r.FullName())
	}
	return names
}

// build a map of releases keyed by full names
func buildReleaseMap(releases []terra.Release) map[string]terra.Release {
	m := make(map[string]terra.Release)
	for _, release := range releases {
		m[release.FullName()] = release
	}
	return m
}

package source

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

const configFile = ".autorelease.yaml"
const targetVersionSet = gitops.Dev

// AutoReleaser offers a UpdateReleaseVersion to take a newly published chart and update development instances to use
// it.
// It stores lists of different update mechanisms so they can be easily enabled/disabled by the caller.
// This is a literal struct, not an interface, so the callers can configure it out without needing to pass multiple
// parameters around.
type AutoReleaser struct {
	GitopsUpdaters           []gitops.Versions
	SherlockUpdaters         []sherlock.ChartVersionUpdater
	SoftFailSherlockUpdaters []sherlock.ChartVersionUpdater
}

// Struct for parsing an autorelease.yaml config file
type config struct {
	Enabled bool `yaml:"enabled"` // whether updates to this chart should be added to release train. defaults to true

	// Release is "legacy" config for chart version autoreleases. It is all that gitops versioning can pay attention
	// to, and sherlock versioning will pay attention to it if its own Sherlock config isn't provided.
	Release struct {
		Name string            `yaml:"name"` // name of the "release", defaults to chart name
		Type terra.ReleaseType `yaml:"type"` // either "app" or "cluster", defaults to app
	} `yaml:"release"`

	// Sherlock is config for chart version autoreleases using the sherlock versioning system.
	//
	// The older Release configuration will still be used to configure the chart name. If this Sherlock config isn't
	// provided at all, there is soft-fail default behavior to try to update the dev instance of the chart.
	Sherlock struct {
		// ChartReleasesToUseLatest is a list of chart release selectors that will be resolved directly by sherlock.
		// If this chart has a new version autoreleased and that new version was successfully reported to sherlock,
		// all chart releases in this list will be set to use sherlock's current latest chart version (which should
		// be the new version, barring reruns or other double-send issues that sherlock will resolve).
		//
		// If this isn't provided, the default behavior takes effect. This field implicitly gets a value of
		// "dev/<chart>" in soft-fail mode, which means "if an instance of this chart exists in the dev environment,
		// set it to use latest."
		ChartReleasesToUseLatest []string `yaml:"chartReleasesToUseLatest"`
	} `yaml:"sherlock"`
}

func (a *AutoReleaser) UpdateReleaseVersion(chart Chart, newVersion string, lastVersion string, description string) error {
	cfg := loadConfig(chart)
	if !cfg.Enabled {
		return nil
	}

	for index, gitopsUpdater := range a.GitopsUpdaters {
		err := gitopsUpdater.
			GetSnapshot(cfg.Release.Type, targetVersionSet).
			UpdateChartVersionIfDefined(cfg.Release.Name, newVersion)
		if err != nil {
			return fmt.Errorf("autorelease error on gitops updater %d: %v", index, err)
		}
	}

	var sherlockTargetChartReleases []string
	var sherlockCanAlwaysSoftFail bool
	if len(cfg.Sherlock.ChartReleasesToUseLatest) > 0 {
		sherlockTargetChartReleases = cfg.Sherlock.ChartReleasesToUseLatest
		sherlockCanAlwaysSoftFail = false
	} else {
		sherlockTargetChartReleases = []string{fmt.Sprintf("%s/%s", targetVersionSet.String(), cfg.Release.Name)}
		sherlockCanAlwaysSoftFail = true
	}
	for index, sherlockUpdater := range a.SherlockUpdaters {
		err := sherlockUpdater.
			UpdateForNewChartVersion(cfg.Release.Name, newVersion, lastVersion, description, sherlockTargetChartReleases...)
		if err != nil {
			if sherlockCanAlwaysSoftFail {
				log.Warn().Err(err).Msgf("autorelease error on sherlock updater %d: %v", index, err)
			} else {
				return fmt.Errorf("autorelease error on sherlock updater %d: %v", index, err)
			}
		}
	}
	for index, sherlockUpdater := range a.SoftFailSherlockUpdaters {
		err := sherlockUpdater.
			UpdateForNewChartVersion(cfg.Release.Name, newVersion, lastVersion, description, sherlockTargetChartReleases...)
		if err != nil {
			log.Debug().Err(err).Msgf("autorelease error on sherlock soft-fail updater %d: %v", index, err)
		}
	}

	return nil
}

// load .autorelease.yaml config file from chart source directory if it exists
func loadConfig(chart Chart) config {
	cfg := config{}

	// Set defaults
	cfg.Enabled = true
	cfg.Release.Name = chart.Name()
	cfg.Release.Type = terra.AppReleaseType

	file := path.Join(chart.Path(), configFile)
	_, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warn().Msgf("unexpected error reading %s: %v, falling back to default config", file, err)
		}
		// no config file or can't read it, so return empty
		return cfg
	}

	data, err := os.ReadFile(file)
	if err != nil {
		log.Warn().Msgf("unexpected error reading %s: %v, falling back to default config", file, err)
		return cfg
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Warn().Msgf("unexpected error parsing %s: %v, falling back to default config", file, err)
		return cfg
	}

	return cfg
}
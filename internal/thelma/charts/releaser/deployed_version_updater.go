package releaser

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

const configFile = ".autorelease.yaml"

const targetEnvironment = "dev"

// DeployedVersionUpdater offers a UpdateReleaseVersion to take a newly published chart and update development instances to use
// it.
// It stores lists of different update mechanisms so they can be easily enabled/disabled by the caller.
// This is a literal struct, not an interface, so the callers can configure it out without needing to pass multiple
// parameters around.
type DeployedVersionUpdater struct {
	SherlockUpdaters         []sherlock.ChartVersionUpdater
	SoftFailSherlockUpdaters []sherlock.ChartVersionUpdater
}

// Struct for parsing an autorelease.yaml config file
type config struct {
	// Enabled defaults to true; set explicitly to false to disable autoreleases, i.e
	// don't automatically update any chart releases for this chart when a new version is published
	Enabled bool `yaml:"enabled"`

	// Sherlock is config for chart version autoreleases using the sherlock versioning system.
	//
	// If this Sherlock config isn't provided at all, there is soft-fail default behavior to try to update the
	// dev instance of the chart.
	Sherlock struct {
		// ChartReleasesToUseLatest is a list of chart release names, like "agora-dev" or "yale-terra-prod".
		//
		// If this chart has a new version autoreleased and that new version was successfully reported to sherlock,
		// all chart releases in this list will be set to use sherlock's current latest chart version (which should
		// be the new version, barring reruns or other double-send issues that sherlock will resolve).
		//
		// If this isn't provided, the default behavior takes effect. This field implicitly gets a value of
		// "<chart>-dev" in soft-fail mode, which means "if an instance of this chart exists in the dev environment,
		// set it to use latest."
		ChartReleasesToUseLatest []string `yaml:"chartReleasesToUseLatest"`
	} `yaml:"sherlock"`
}

func (a *DeployedVersionUpdater) UpdateReleaseVersion(chart source.Chart, newVersion string, lastVersion string, description string) ([]string, error) {
	cfg := loadConfig(chart)
	if !cfg.Enabled {
		return nil, nil
	}

	var sherlockTargetChartReleases []string
	var sherlockCanAlwaysSoftFail bool
	if len(cfg.Sherlock.ChartReleasesToUseLatest) > 0 {
		sherlockTargetChartReleases = cfg.Sherlock.ChartReleasesToUseLatest
		sherlockCanAlwaysSoftFail = false
	} else {
		sherlockTargetChartReleases = []string{fmt.Sprintf("%s-%s", chart.Name(), targetEnvironment)}
		sherlockCanAlwaysSoftFail = true
	}
	for index, sherlockUpdater := range a.SherlockUpdaters {
		err := sherlockUpdater.
			UpdateForNewChartVersion(chart.Name(), newVersion, lastVersion, description, sherlockTargetChartReleases...)
		if err != nil {
			if sherlockCanAlwaysSoftFail {
				log.Warn().Err(err).Msgf("autorelease error on sherlock updater %d: %v", index, err)
			} else {
				return nil, errors.Errorf("autorelease error on sherlock updater %d: %v", index, err)
			}
		}
	}
	for index, sherlockUpdater := range a.SoftFailSherlockUpdaters {
		err := sherlockUpdater.
			UpdateForNewChartVersion(chart.Name(), newVersion, lastVersion, description, sherlockTargetChartReleases...)
		if err != nil {
			log.Debug().Err(err).Msgf("autorelease error on sherlock soft-fail updater %d: %v", index, err)
		}
	}

	return sherlockTargetChartReleases, nil
}

// load .autorelease.yaml config file from chart source directory if it exists
func loadConfig(chart source.Chart) config {
	cfg := config{}

	// Set defaults
	cfg.Enabled = true

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

package deploy

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

const configFile = ".autorelease.yaml"

const defaultTargetEnvironment = "dev"

type ConfigLoader interface {
	FindReleasesToUpdate(chartName string) ([]terra.Release, error)
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

func newConfigLoader(chartsDir source.ChartsDir, state terra.State) (ConfigLoader, error) {
	releases, err := state.Releases().All()
	if err != nil {
		return nil, err
	}

	return &configLoaderImpl{
		chartsDir: chartsDir,
		releases:  buildReleaseMap(releases),
	}, nil
}

type configLoaderImpl struct {
	releases  map[string]terra.Release
	chartsDir source.ChartsDir
}

func (m *configLoaderImpl) FindReleasesToUpdate(chartName string) ([]terra.Release, error) {
	chart, err := m.chartsDir.GetChart(chartName)
	if err != nil {
		return nil, errors.Errorf("error loading chart %s from %s: %v", chartName, m.chartsDir.Path(), err)
	}

	cfg := loadConfig(chart)
	if !cfg.Enabled {
		log.Warn().Msgf("autorelease disabled for chart %s, won't attempt a dev deploy", chart.Name())
		return nil, nil
	}

	var releaseNames []string
	if len(cfg.Sherlock.ChartReleasesToUseLatest) > 0 {
		releaseNames = cfg.Sherlock.ChartReleasesToUseLatest
	} else {
		releaseNames = []string{chartName + "-" + defaultTargetEnvironment}
	}

	var releases []terra.Release
	for _, name := range releaseNames {
		release, exists := m.releases[name]
		if !exists {
			if len(cfg.Sherlock.ChartReleasesToUseLatest) > 0 {
				log.Warn().Msgf("chart release %s not found in terra state, won't try to update", name)
			} else {
				log.Debug().Msgf("chart release %s not found in terra state, won't try to update", name)
			}
			continue
		}
		releases = append(releases, release)
	}

	return releases, nil
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

// Package releaser is the main orchestrator for releasing new charts. It can be thought of as the main entrypoint
// all the other logic in the charts/ package.
package releaser

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/publish"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

// ChartReleaser is the main orchestrator for releasing new charts.
type ChartReleaser interface {
	// Release calculates out downstream dependents of the given charts, increments versions,
	// and publishes new chart packages to the Helm repo.
	//
	// Note that Release will release downstream dependents of the charts it is given. In other words, if chart `bar`
	// depends on chart `foo`, just including `foo` in the chartNames will also publish and release `bar`.
	//
	// Params:
	// chartsToPublish: a list of chart names that should be published. Eg. ["foo"]
	// changeDescription: freeform text used to annotate the chart version in Sherlock. Usually corresponds to PR commit message.
	//
	// Return:
	// a map representing the names and versions of charts that were published and released. Eg.
	// {
	//   "foo": "1.2.3",
	//   "bar": "0.2.0",
	// }
	//
	Release(chartsToPublish []string, changeDescription string) (publishedVersions map[string]VersionPair, err error)
}

func NewChartReleaser(chartsDir source.ChartsDir, publisher publish.Publisher, deployedVersionUpdater *DeployedVersionUpdater) ChartReleaser {
	return &chartReleaser{
		chartsDir:              chartsDir,
		publisher:              publisher,
		deployedVersionUpdater: deployedVersionUpdater,
	}
}

type chartReleaser struct {
	chartsDir              source.ChartsDir
	publisher              publish.Publisher
	deployedVersionUpdater *DeployedVersionUpdater
}

type VersionPair struct {
	// version of the chart directly preceding NewVersion
	PriorVersion string
	// the new version of the chart that will be published
	NewVersion string
}

func (r *chartReleaser) Release(chartNames []string, description string) (map[string]VersionPair, error) {
	// make sure all charts exist in source dir
	chartsToPublish := chartNames
	for _, chartName := range chartsToPublish {
		if !r.chartsDir.Exists(chartName) {
			return nil, errors.Errorf("chart %s does not exist in source directory", chartName)
		}
	}

	// add dependents
	withDependents, err := r.withDependents(chartsToPublish)
	if err != nil {
		return nil, err
	}
	chartsToPublish = withDependents

	log.Info().Msgf("%d charts will be published: %s", len(chartsToPublish), strings.Join(chartsToPublish, ", "))

	// identify new version for each chart and bump in Chart.yaml
	chartVersions, err := r.bumpChartVersions(chartsToPublish)
	if err != nil {
		return nil, err
	}

	// run `helm dependency update` on all charts we're publishing, plus their transitive dependencies, in
	// topological order
	if err = r.updateAllDependencies(chartsToPublish); err != nil {
		return nil, err
	}

	// generate docs and package charts in the publisher's staging directory
	if err = r.packageCharts(chartsToPublish); err != nil {
		return nil, err
	}

	// upload charts to helm repo
	if err = r.publishCharts(); err != nil {
		return nil, err
	}

	// report new chart versions to Sherlock
	if err = r.reportNewVersionsToSherlock(chartVersions, description); err != nil {
		return nil, err
	}

	return chartVersions, nil
}

func (r *chartReleaser) reportNewVersionsToSherlock(chartVersions map[string]VersionPair, description string) error {
	for chartName, versions := range chartVersions {
		if err := r.deployedVersionUpdater.ReportNewChartVersion(chartName, versions, description); err != nil {
			return err
		}
	}
	log.Info().Msgf("%d new chart versions were reported to Sherlock", len(chartVersions))
	return nil
}

func (r *chartReleaser) publishCharts() error {
	count, err := r.publisher.Publish()
	if err != nil {
		return err
	}
	log.Info().Msgf("%d charts were uploaded to the repository", count)
	return nil
}

func (r *chartReleaser) packageCharts(chartNames []string) error {
	for _, chartName := range chartNames {
		_chart, err := r.chartsDir.GetChart(chartName)
		if err != nil {
			return err
		}

		if err = _chart.GenerateDocs(); err != nil {
			return err
		}

		if err = _chart.PackageChart(r.publisher.ChartDir()); err != nil {
			return err
		}
	}

	return nil
}

func (r *chartReleaser) updateAllDependencies(chartNames []string) error {
	charts, err := r.chartsDir.GetCharts(chartNames...)
	if err != nil {
		return err
	}
	return r.chartsDir.RecursivelyUpdateDependencies(charts...)
}

func (r *chartReleaser) bumpChartVersions(chartNames []string) (map[string]VersionPair, error) {
	chartVersions := make(map[string]VersionPair, len(chartNames))
	for _, chartName := range chartNames {
		releaseInfo, err := r.bumpChartVersion(chartName)
		if err != nil {
			return nil, err
		}
		chartVersions[chartName] = *releaseInfo
	}
	return chartVersions, nil
}

func (r *chartReleaser) bumpChartVersion(chartName string) (*VersionPair, error) {
	_chart, err := r.chartsDir.GetChart(chartName)
	if err != nil {
		return nil, err
	}

	// get last version of chart
	lastVersion := r.publisher.Index().MostRecentVersion(chartName)
	if err != nil {
		return nil, err
	}

	// bump chart version in Chart.yaml
	newVersion, err := _chart.BumpChartVersion(lastVersion)
	if err != nil {
		return nil, err
	}

	// for all charts that depend on this chart, update their Chart.yaml files to use
	// the new version of this chart
	if err = r.chartsDir.UpdateDependentVersionConstraints(_chart, newVersion); err != nil {
		return nil, err
	}

	return &VersionPair{
		PriorVersion: lastVersion,
		NewVersion:   newVersion,
	}, nil
}

// given a list of chart names, return the list of charts with dependents included, sorted topologically.
// eg.
// if "tps" -> "foundation" -> "ingress", then
// withDependents("ingress") returns
// ["ingress", "foundation", "tps"]
func (r *chartReleaser) withDependents(chartNames []string) ([]string, error) {
	asCharts, err := r.chartsDir.GetCharts(chartNames...)
	if err != nil {
		return nil, err
	}

	withDependents, err := r.chartsDir.WithTransitiveDependents(asCharts)
	if err != nil {
		return nil, err
	}

	return source.ChartNames(withDependents...), nil
}

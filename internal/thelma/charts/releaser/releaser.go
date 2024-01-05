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

const maxParallelSync = 30

// ChartReleaser is the main orchestrator for releasing new charts.
type ChartReleaser interface {
	// Release calculates out downstream dependents of the given charts, increments versions, publishes new
	// chart packages to the Helm repo, and releases those new versions into Sherlock.
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
	Release(chartsToPublish []string, changeDescription string) (publishedVersions map[string]string, err error)
}

func NewChartReleaser(chartsDir source.ChartsDir, publisher publish.Publisher, updater *DeployedVersionUpdater, syncer PostUpdateSyncer) ChartReleaser {
	return &chartReleaser{
		chartsDir: chartsDir,
		publisher: publisher,
		updater:   updater,
		syncer:    syncer,
	}
}

// okay so - maybe we want an abstraction here for syncing the releases.
// post_deploy_syncer - accepts a dry run parameter and clients.
type chartReleaser struct {
	chartsDir source.ChartsDir
	publisher publish.Publisher
	updater   *DeployedVersionUpdater
	syncer    PostUpdateSyncer
}

type versions struct {
	// previous version of the chart
	lastVersion string
	// new version of the chart that will be published
	newVersion string
}

func (r *chartReleaser) Release(chartNames []string, description string) (map[string]string, error) {
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

	// We run the updater after publishing the charts to avoid an instance where a chart release points at a chart
	// version that hasn't been published quite yet
	publishedVersions, updatedChartReleaseNames, err := r.reportNewChartVersionsToSherlock(chartVersions, description)
	if err != nil {
		return nil, err
	}

	// sync updated chart releases
	if err = r.syncer.Sync(updatedChartReleaseNames); err != nil {
		return nil, err
	}

	return publishedVersions, nil
}

// given a map of chart version info, reportNewChartVersionsToSherlock will report the new versions to Sherlock.
//
// Return:
// a map representing the names and versions of charts that were published and released. Eg.
//
//	{
//	  "foo": "1.2.3",
//	  "bar": "0.2.0",
//	}
func (r *chartReleaser) reportNewChartVersionsToSherlock(chartVersions map[string]versions, description string) (map[string]string, []string, error) {
	publishedVersions := make(map[string]string, len(chartVersions))

	var updatedReleases []string

	if r.updater != nil {
		for chartName, versions := range chartVersions {
			chart, err := r.chartsDir.GetChart(chartName)
			if err != nil {
				return nil, nil, err
			}
			updated, err := r.updater.UpdateReleaseVersion(chart, versions.newVersion, versions.lastVersion, description)
			if err != nil {
				return publishedVersions, updatedReleases, err
			}
			publishedVersions[chartName] = versions.newVersion
			updatedReleases = append(updatedReleases, updated...)
		}
	}

	log.Info().Msgf("Attempted to update %d chart releases: %s", len(updatedReleases), strings.Join(updatedReleases, ", "))

	return publishedVersions, updatedReleases, nil
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

func (r *chartReleaser) bumpChartVersions(chartNames []string) (map[string]versions, error) {
	chartVersions := make(map[string]versions, len(chartNames))
	for _, chartName := range chartNames {
		releaseInfo, err := r.bumpChartVersion(chartName)
		if err != nil {
			return nil, err
		}
		chartVersions[chartName] = *releaseInfo
	}
	return chartVersions, nil
}

func (r *chartReleaser) bumpChartVersion(chartName string) (*versions, error) {
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

	return &versions{
		lastVersion: lastVersion,
		newVersion:  newVersion,
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

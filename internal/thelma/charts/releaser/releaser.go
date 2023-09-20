// Package releaser is the main orchestrator for releasing new charts. It can be thought of as the main entrypoint
// all the other logic in the charts/ package.
package releaser

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/publish"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/rs/zerolog/log"
	"strings"
)

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

func NewChartReleaser(sourceDir source.ChartsDir, publisher publish.Publisher, updater *DeployedVersionUpdater) ChartReleaser {
	return &chartReleaser{
		sourceDir: sourceDir,
		publisher: publisher,
		updater:   updater,
	}
}

type chartReleaser struct {
	sourceDir source.ChartsDir
	publisher publish.Publisher
	updater   *DeployedVersionUpdater
}

func (r *chartReleaser) Release(chartNames []string, description string) (map[string]string, error) {
	chartsToPublish := chartNames
	for _, chartName := range chartsToPublish {
		_, err := r.sourceDir.GetChart(chartName)
		if err != nil {
			return nil, err
		}
	}

	log.Info().Msgf("%d charts will be published: %s", len(chartsToPublish), strings.Join(chartsToPublish, ", "))

	publishedVersions := make(map[string]string, len(chartsToPublish))
	lastVersions := make(map[string]string, len(chartsToPublish))
	for _, chartName := range chartsToPublish {
		_chart, err := r.sourceDir.GetChart(chartName)
		if err != nil {
			return nil, err
		}

		dependenciesToUpdate := r.sourceDir.DetermineDependenciesToUpdate(_chart)

		lastVersions[chartName] = r.publisher.Index().MostRecentVersion(chartName)
		newVersion, err := _chart.BumpChartVersion(lastVersions[chartName])
		if err != nil {
			return nil, err
		}
		if err := r.sourceDir.UpdateDependentVersionConstraints(_chart, newVersion); err != nil {
			return nil, err
		}
		for _, chartToUpdate := range dependenciesToUpdate {
			_chartToUpdate, err := r.sourceDir.GetChart(chartToUpdate)
			if err != nil {
				return nil, err
			}
			if err := _chartToUpdate.UpdateDependencies(); err != nil {
				return nil, err
			}
		}

		if err := _chart.GenerateDocs(); err != nil {
			return nil, err
		}

		if err := _chart.PackageChart(r.publisher.ChartDir()); err != nil {
			return nil, err
		}

		publishedVersions[chartName] = newVersion
	}

	count, err := r.publisher.Publish()
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("%d charts were uploaded to the repository", count)

	// We run the updater after publishing the charts to avoid an instance where a chart release points at a chart
	// version that hasn't been published quite yet
	if r.updater != nil {
		for _, chartName := range chartsToPublish {
			chart, err := r.sourceDir.GetChart(chartName)
			if err != nil {
				return nil, err
			}
			err = r.updater.UpdateReleaseVersion(chart, publishedVersions[chartName], lastVersions[chartName], description)
			if err != nil {
				return publishedVersions, err
			}
		}
	}

	return publishedVersions, nil
}

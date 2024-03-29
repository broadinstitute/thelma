package resolver

import (
	"github.com/pkg/errors"
	"os"
	"path"

	"github.com/broadinstitute/thelma/internal/thelma/charts/dependency"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

// LocalResolver is for resolving charts in local directory of chart sources.
// (i.e. the charts/ subdirectory in terra-helmfile).
type localResolver interface {
	// True if the chart exists in the source directory, else false.
	// Returns error in the event of an unexpected filesystem i/o error
	chartExists(chart ChartRelease) (bool, error)
	// Locates the chart in the source directory, running `helm dependency update` on it if needed
	resolve(chart ChartRelease) (ResolvedChart, error)
	// Returns version of the chart in the chart's Chart.yaml
	sourceVersion(chart ChartRelease) (string, error)
}

type localResolverImpl struct {
	sourceDir string
	cache     syncCache[source.Chart]
	runner    shell.Runner
}

func newLocalResolver(sourceDir string, runner shell.Runner) localResolver {
	r := &localResolverImpl{
		sourceDir: sourceDir,
		runner:    runner,
	}
	r.cache = newSyncCacheWithMapper(r.resolverFn, func(resolvable source.Chart) string {
		return resolvable.Name()
	})
	return r
}

func (r *localResolverImpl) chartExists(chartRelease ChartRelease) (bool, error) {
	chartPath := r.chartSourcePath(chartRelease.Name)
	_, err := os.Stat(chartPath)

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, errors.Errorf("error checking for chart at %s: %v", chartPath, err)
		}
	}

	return true, nil
}

func (r *localResolverImpl) sourceVersion(chartRelease ChartRelease) (string, error) {
	chart, err := r.getChart(chartRelease.Name)
	if err != nil {
		return "", errors.Errorf("error reading chart %s: %v", chartRelease.Name, err)
	}

	return chart.ManifestVersion(), nil
}

// resolve gets a ResolvedChart for the given ChartRelease, but it behaves differently under the hood compared
// to its counterparts in remoteResolverImpl and the top-level chartResolver.
//
// The resolve methods in those other implementations just hand off to their respective syncCache, and all the logic
// is inside the resolver function inside the syncCache. That works because there's a unique process to run for
// each ChartRelease (say, downloading the appropriately versioned tarball) and that process is thread-safe.
//
// That won't work here, though. For one, when we're resolving locally, we don't care about the chart *release*, we
// just care about the chart (the version / contents are derived from the filesystem, so it's the same regardless of
// what environment the chart is deployed to). That means we can have our syncCache operate per chart, not per
// chart release.
//
// Secondly, `helm dependency update` invocations against the same directory aren't thread-safe. This means we need to
// prevent concurrent invocations against not just top-level charts, but the other dependent charts on disk.
//
// syncCache still works, but rather than caching per ChartRelease, we instead do caching per source.Chart, including
// dependencies courtesy of determineChartDependencies. We still only return the ResolvedChart that the caller cares
// about for their input ChartRelease, but caching for all the underlying dependencies gets us the locking behavior we
// need.
func (r *localResolverImpl) resolve(chartRelease ChartRelease) (ResolvedChart, error) {
	chart, err := r.getChart(chartRelease.Name)
	if err != nil {
		return nil, err
	}
	dependencyCharts, err := r.determineChartDependencies(chart)
	if err != nil {
		return nil, errors.Errorf("error determining dependencies of chart %s: %v", chart.Name(), err)
	}
	var desiredResolvedChart ResolvedChart
	for _, dependencyChart := range dependencyCharts {
		resolvedDependencyChart, err := r.cache.get(dependencyChart)
		if err != nil {
			return nil, errors.Errorf("error resolving chart %s in %s, necessary for chart %s: %v", dependencyChart.Name(), dependencyChart.Path(), chart.Name(), err)
		}
		if dependencyChart.Name() == chart.Name() {
			desiredResolvedChart = resolvedDependencyChart
		}
	}
	return desiredResolvedChart, nil
}

func (r *localResolverImpl) resolverFn(chart source.Chart) (ResolvedChart, error) {
	if err := chart.UpdateDependencies(); err != nil {
		return nil, errors.Errorf("error updating chart source directory %s: %v", chart.Path(), err)
	}
	return NewLocallyResolvedChart(chart.Path(), chart.ManifestVersion()), nil
}

// Returns source.Chart instance for the given chart name
func (r *localResolverImpl) getChart(chartName string) (source.Chart, error) {
	chart, err := source.NewChart(r.chartSourcePath(chartName), r.runner)
	if err != nil {
		return nil, errors.Errorf("error reading chart %s in %s: %v", chartName, r.sourceDir, err)
	}
	return chart, nil
}

// Returns the path to chart source directory
func (r *localResolverImpl) chartSourcePath(chartName string) string {
	return path.Join(r.sourceDir, chartName)
}

// determineChartDependencies is used to add support for lack of handling for multi-layer dependencies
// in helm upgrade. It performs a BFS traversal of the dependency graph for a given chart and outputs
// a topologically sorted list of charts to run helm dependency update upon
func (r *localResolverImpl) determineChartDependencies(chart source.Chart) ([]source.Chart, error) {
	dependencies := make(map[string][]string)
	dependencies[chart.Name()] = chart.LocalDependencies()
	chartsToProcess := make([]string, 0)
	chartsToProcess = append(chartsToProcess, chart.LocalDependencies()...)

	// BFS tree traversal of dependencies for a single chart
	for len(chartsToProcess) != 0 {
		currentChartName := chartsToProcess[0]
		chartsToProcess = chartsToProcess[1:]

		currentChart, err := r.getChart(currentChartName)
		if err != nil {
			return nil, errors.Errorf(
				"error finding chart source locally while calculating depedencies. parent chart: %s, %v",
				chart.Name(), err,
			)
		}

		dependencies[currentChart.Name()] = currentChart.LocalDependencies()
		chartsToProcess = append(chartsToProcess, currentChart.LocalDependencies()...)
	}

	dependencyGraph, err := dependency.NewGraph(dependencies)
	if err != nil {
		return nil, errors.Errorf("error constructing depdency graph in local resolver: %v", err)
	}

	dependenciesToUpdate := make([]string, 0)
	// dependency map keys need to be transferred into a slice in order to use Topo sort
	for chart := range dependencies {
		dependenciesToUpdate = append(dependenciesToUpdate, chart)
	}
	dependencyGraph.TopoSort(dependenciesToUpdate)

	// convert to domain type
	var chartsToUpdate []source.Chart
	for _, chart := range dependenciesToUpdate {
		sourceChart, err := r.getChart(chart)
		if err != nil {
			return nil, errors.Errorf("error finding local source for chart: %s, %v", chart, err)
		}
		chartsToUpdate = append(chartsToUpdate, sourceChart)
	}
	return chartsToUpdate, nil
}

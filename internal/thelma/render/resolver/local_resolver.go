package resolver

import (
	"fmt"
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
	cache     syncCache
	runner    shell.Runner
}

func newLocalResolver(sourceDir string, runner shell.Runner) localResolver {
	cache := newSyncCacheWithMapper(func(chart ChartRelease) string {
		// since all charts live at the first level of the source directory
		// eg. charts/agora, charts/cromwell, etc.
		// we key on chart name and ignore version and repository
		return chart.Name
	})

	return &localResolverImpl{
		sourceDir: sourceDir,
		cache:     cache,
		runner:    runner,
	}
}

func (r *localResolverImpl) chartExists(chartRelease ChartRelease) (bool, error) {
	chartPath := r.chartSourcePath(chartRelease.Name)
	_, err := os.Stat(chartPath)

	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("error checking for chart at %s: %v", chartPath, err)
		}
	}

	return true, nil
}

func (r *localResolverImpl) sourceVersion(chartRelease ChartRelease) (string, error) {
	chart, err := r.getChart(chartRelease.Name)
	if err != nil {
		return "", fmt.Errorf("error reading chart %s: %v", chartRelease.Name, err)
	}

	return chart.ManifestVersion(), nil
}

func (r *localResolverImpl) resolve(chartRelease ChartRelease) (ResolvedChart, error) {
	return r.cache.get(chartRelease, r.resolverFn)
}

func (r *localResolverImpl) resolverFn(chartRelease ChartRelease) (ResolvedChart, error) {
	chart, err := r.getChart(chartRelease.Name)
	if err != nil {
		return nil, err
	}
	chartsToUpdate, err := r.determineDependenciesToUpdate(chart)
	if err != nil {
		return nil, fmt.Errorf("error determining charts to run helm dependency update on: %v", err)
	}
	for _, chart := range chartsToUpdate {
		err = chart.UpdateDependencies()
		if err != nil {
			return nil, fmt.Errorf("error updating chart source directory %s: %v", chart.Path(), err)
		}
	}

	return NewResolvedChart(chart.Path(), chart.ManifestVersion(), Local, chartRelease), nil
}

// Returns source.Chart instance for the given chart name
func (r *localResolverImpl) getChart(chartName string) (source.Chart, error) {
	chart, err := source.NewChart(r.chartSourcePath(chartName), r.runner)
	if err != nil {
		return nil, fmt.Errorf("error reading chart %s in %s: %v", chartName, r.sourceDir, err)
	}
	return chart, nil
}

// Returns the path to chart source directory
func (r *localResolverImpl) chartSourcePath(chartName string) string {
	return path.Join(r.sourceDir, chartName)
}

// determineDependenciesToUpdate is used to add support for lack of handling for multi-layer dependencies
// in helm upgrade. It performs a BFS traversal of the dependency graph for a given chart and outputs
// a topologically sorted list of charts to run helm dependency update upon
func (r *localResolverImpl) determineDependenciesToUpdate(chart source.Chart) ([]source.Chart, error) {
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
			return nil, fmt.Errorf(
				"error finding chart source locally while calculating depedencies. parent chart: %s, %v",
				chart.Name(), err,
			)
		}

		dependencies[currentChart.Name()] = currentChart.LocalDependencies()
		chartsToProcess = append(chartsToProcess, currentChart.LocalDependencies()...)
	}

	dependencyGraph, err := dependency.NewGraph(dependencies)
	if err != nil {
		return nil, fmt.Errorf("error constructing depdency graph in local resolver: %v", err)
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
			return nil, fmt.Errorf("error finding local source for chart: %s, %v", chart, err)
		}
		chartsToUpdate = append(chartsToUpdate, sourceChart)
	}
	return chartsToUpdate, nil
}

package source

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/dependency"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"path"
	"path/filepath"
)

// ChartsDir represents a directory of Helm chart sources on the local filesystem.
type ChartsDir interface {
	// Path returns the filesystem path of the source directory
	Path() string
	// Exists returns true if a chart by the given name exists in the directory
	Exists(name string) bool
	// GetChart returns a Chart for the chart with the given name, or an error if no chart by that name exists in source dir
	GetChart(name string) (Chart, error)
	// GetCharts returns the Chart objects for a set of charts with the given name(s), or an error if no chart(s) by that name exists in source dir
	GetCharts(name ...string) ([]Chart, error)
	// UpdateDependentVersionConstraints go through all dependents and update version constraints to match the new one
	UpdateDependentVersionConstraints(chart Chart, newVersionConstraint string) error
	// WithTransitiveDependents returns the given charts plus all of their transitive dependents
	WithTransitiveDependents(chart []Chart) ([]Chart, error)
	// RecursivelyUpdateDependencies given a list of charts:
	// * identify all transitive dependencies of those charts
	// * sort the list of charts + dependencies in topological order
	// * run `helm dependency update` on each chart in the list
	RecursivelyUpdateDependencies(chart ...Chart) error
}

// NewChartsDir constructs a new ChartsDir
func NewChartsDir(
	sourceDir string,
	shellRunner shell.Runner,
) (ChartsDir, error) {
	charts, err := loadCharts(sourceDir, shellRunner)
	if err != nil {
		return nil, err
	}

	dependencyGraph, err := buildDependencyGraph(charts)
	if err != nil {
		return nil, err
	}

	return &chartsDir{
		sourceDir:       sourceDir,
		charts:          charts,
		dependencyGraph: dependencyGraph,
	}, nil
}

func ChartNames(charts ...Chart) []string {
	var names []string
	for _, c := range charts {
		names = append(names, c.Name())
	}
	return names
}

// implemeents ChartsDir interface
type chartsDir struct {
	sourceDir       string
	charts          map[string]Chart
	dependencyGraph *dependency.Graph
}

func (d *chartsDir) Path() string {
	return d.sourceDir
}

func (d *chartsDir) Exists(chartName string) bool {
	_, exists := d.charts[chartName]
	return exists
}

func (d *chartsDir) GetChart(chartName string) (Chart, error) {
	_chart, exists := d.charts[chartName]

	if !exists {
		return nil, errors.Errorf("chart %q does not exist in source dir %s", chartName, d.sourceDir)
	}
	return _chart, nil
}

func (d *chartsDir) GetCharts(chartNames ...string) ([]Chart, error) {
	var charts []Chart
	for _, name := range chartNames {
		_chart, err := d.GetChart(name)
		if err != nil {
			return nil, err
		}
		charts = append(charts, _chart)
	}

	return charts, nil
}

// UpdateDependentVersionConstraints go through all dependents and update version constraints to match the new one
func (d *chartsDir) UpdateDependentVersionConstraints(chart Chart, newVersionConstraint string) error {
	for _, dependent := range d.dependencyGraph.GetDependents(chart.Name()) {
		dependentChart := d.charts[dependent]
		if err := dependentChart.SetDependencyVersion(chart.Name(), newVersionConstraint); err != nil {
			return err
		}
	}
	return nil
}

// WithTransitiveDependents returns the given charts plus all of their transitive dependents, in topologically sorted order
func (d *chartsDir) WithTransitiveDependents(charts []Chart) ([]Chart, error) {
	chartNames := ChartNames(charts...)

	result := d.dependencyGraph.WithTransitiveDependents(chartNames...)

	diff := len(result) - len(chartNames)
	if diff > 0 {
		log.Info().Msgf("Identified %d additional downstream charts to publish", diff)
	}

	d.dependencyGraph.TopoSort(result)

	return d.GetCharts(result...)
}

// RecursivelyUpdateDependencies given a list of charts:
// * identify all transitive dependencies of those charts
// * sort the list of charts + dependencies in topological order
// * run `helm dependency update` on each chart in the list
func (d *chartsDir) RecursivelyUpdateDependencies(charts ...Chart) error {
	withDependenciesNames, err := d.findTransitiveDependencies(charts...)
	if err != nil {
		return err
	}

	withDependencies, err := d.GetCharts(withDependenciesNames...)
	if err != nil {
		return err
	}

	for _, _chart := range withDependencies {
		if err = _chart.UpdateDependencies(); err != nil {
			return err
		}
	}

	return nil
}

// findTransitiveDependencies returns the names of all charts in the dependency trees of the given chart(s)
// in topologically sorted order
func (d *chartsDir) findTransitiveDependencies(charts ...Chart) ([]string, error) {
	var chartsToProcess []string
	visited := set.NewSet[string]()

	for _, chart := range charts {
		chartsToProcess = append(chartsToProcess, chart.Name())
	}

	for len(chartsToProcess) != 0 {
		currentChartName := chartsToProcess[0]
		chartsToProcess = chartsToProcess[1:]

		visited.Add(currentChartName)

		currentChart, err := d.GetChart(currentChartName)
		if err != nil {
			return nil, errors.Errorf("error processing chart %q: %v", currentChartName, err)
		}

		for _, dep := range currentChart.LocalDependencies() {
			if !visited.Exists(dep) {
				chartsToProcess = append(chartsToProcess, dep)
			}
		}
	}

	dependenciesToUpdate := visited.Elements()
	d.dependencyGraph.TopoSort(dependenciesToUpdate)

	return dependenciesToUpdate, nil
}

func buildDependencyGraph(charts map[string]Chart) (*dependency.Graph, error) {
	dependencies := make(map[string][]string)
	for chartName, _chart := range charts {
		var localDeps []string
		for _, depName := range _chart.LocalDependencies() {
			// double-check that the dependencies actually exist in the chart dir
			if _, exists := charts[depName]; !exists {
				log.Warn().Msgf("chart %s dependency %s is not in source dir, ignoring", _chart.Name(), depName)
				continue
			}
			localDeps = append(localDeps, depName)
		}
		dependencies[chartName] = localDeps
	}

	return dependency.NewGraph(dependencies)
}

func loadCharts(sourceDir string, shellRunner shell.Runner) (map[string]Chart, error) {
	// Glob inside the chart source directory for chart.yaml files
	glob := path.Join(sourceDir, path.Join("*", chartManifestFile))
	manifestFiles, err := filepath.Glob(glob)
	if err != nil {
		return nil, errors.Errorf("error globbing charts with %q: %v", glob, err)
	}

	// For each chart.yaml file, parse it and store in collection of chart objects
	charts := make(map[string]Chart)

	for _, manifestFile := range manifestFiles {
		// Create node for this chart
		_chart, err := NewChart(path.Dir(manifestFile), shellRunner)
		if err != nil {
			return nil, errors.Errorf("error creating chart from %s: %v", manifestFile, err)
		}
		charts[_chart.Name()] = _chart
	}

	return charts, nil
}

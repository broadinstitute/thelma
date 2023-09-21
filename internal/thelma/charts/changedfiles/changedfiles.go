// Package changedfiles maps a list of updated files in the terra-helmfile repo to a list of charts that need to be published.
//
// The logic is as follows:
// * Changes to charts/<chartname>/* or values/(app|cluster)/<chartname>* will trigger a publish of the affected chart
// * Changes to values/app/global* will trigger a publish of all app release charts
// * Changes to values/cluster/global* will trigger a publish of all cluster release charts
// * Changes to helmfile.yaml will trigger a publish/render of all charts that have at least one chart release
// * Finally, any charts that are transitive dependencies of charts in the above list will be published as well
package changedfiles

import (
	"bufio"
	"bytes"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

const FlagName = "changed-files-list"

type ChangedFiles interface {
	// ChartList returns a list of charts that need to be published/released based on the list of changed files
	//
	// Note that inputFile should be path to a file that contains a newline-separated list of files
	// that were updated by a PR.
	//
	// All paths in the file should be relative to the root of the terra-helmfile repo.
	//
	// Example contents:
	//   charts/agora/templates/deployment.yaml
	//   helmfile.yaml
	//   values/cluster/yale/terra.yaml
	//
	// Note that charts that depend on global values, but don't exist in the chart source directory (i.e., datarepo),
	// are excluded from the list.
	ChartList(inputFile string) ([]string, error)
	// ReleaseFilter is like ChartList, except it returns a filter that matches all terra.Release instances that
	// use a chart that would be published, based on the given list of changed files
	//
	// Note that releases that depend on global values, but don't exist in the chart source directory (i.e., datarepo),
	// will be included by the filter.
	ReleaseFilter(inputFile string) (terra.ReleaseFilter, error)
}

func New(chartsDir source.ChartsDir, state terra.State) ChangedFiles {
	return &changedFiles{
		chartsDir: chartsDir,
		state:     state,
	}
}

type changedFiles struct {
	chartsDir source.ChartsDir
	state     terra.State
}

// globalFileMatches used internally to track which global files were updated
type globalFileMatches struct {
	// include releases that belong to an environment
	includeEnvReleases bool
	// include cluster releases (releases that don't belong to an environment)
	includeClusterReleases bool
	// include all releases
	includeAllReleases bool
}

func (c *changedFiles) ChartList(inputFile string) ([]string, error) {
	chartList, err := c.identifyImpactedCharts(inputFile)
	if err != nil {
		return nil, err
	}

	// filter out any charts that don't exist in the chart source directory
	var exist []string
	for _, chartName := range chartList.Elements() {
		if c.chartsDir.Exists(chartName) {
			exist = append(exist, chartName)
		}
	}

	sort.Strings(exist)
	return exist, nil
}

func (c *changedFiles) ReleaseFilter(inputFile string) (terra.ReleaseFilter, error) {
	charts, err := c.identifyImpactedCharts(inputFile)
	if err != nil {
		return nil, err
	}
	return filter.Releases().HasChartName(charts.Elements()...), nil
}

// identifyImpactedCharts will build a list of charts that are impacted by updated files.
// this includes:
// * charts with chart or values files that have changed
// * charts for releases that are impacted by changes to global values files
// * and the transitive dependents of the above.
// note that the result can include the names of charts that do not exist in the chart
// source directory (for example, datarepo).
func (c *changedFiles) identifyImpactedCharts(inputFile string) (set.Set[string], error) {
	updatedFiles, err := parseChangedList(inputFile)
	if err != nil {
		return nil, errors.Errorf("error parsing %s: %v", inputFile, err)
	}

	var matches globalFileMatches
	chartNames := set.NewSet[string]()

	for _, updatedFileName := range updatedFiles {
		// remove . and .. components in path
		cleaned := path.Clean(updatedFileName)

		// loop through files and update filter options based on which files are included
		if regexp.MustCompile("^charts/[^/]").MatchString(cleaned) {
			chartName := pathIndex(cleaned, 1)
			chartNames.Add(chartName)

		} else if regexp.MustCompile("^values/(app|cluster)/global").MatchString(cleaned) {
			kind := pathIndex(cleaned, 1)
			if kind == "app" {
				matches.includeEnvReleases = true
			} else {
				matches.includeClusterReleases = true
			}

		} else if regexp.MustCompile("^values/[^/]").MatchString(cleaned) {
			component := pathIndex(cleaned, 2)
			// remove any . extensions like .yaml, .yaml.gotmpl
			chartName := strings.Split(component, ".")[0]
			chartNames.Add(chartName)

		} else if regexp.MustCompile("^helmfile.yaml$").MatchString(cleaned) {
			matches.includeAllReleases = true
		}
	}

	// identify releases that are impacted by global values files, and add all their charts to the list
	_filter := matchesToReleaseFilter(matches)
	matchingReleases, err := c.state.Releases().Filter(_filter)
	if err != nil {
		return nil, err
	}

	for _, r := range matchingReleases {
		chartNames.Add(r.ChartName())
	}

	// for any charts currently in the list, add their transitive dependents.
	// this means that if an upstream common dependency chart like foundation or ingress is updated,
	// all downstream charts will be included as well
	if err = c.addDependents(chartNames); err != nil {
		return nil, err
	}

	return chartNames, nil
}

func (c *changedFiles) addDependents(chartNames set.Set[string]) error {
	var exists []string
	for _, chartName := range chartNames.Elements() {
		if c.chartsDir.Exists(chartName) {
			exists = append(exists, chartName)
		}
	}

	asCharts, err := c.chartsDir.GetCharts(exists...)
	if err != nil {
		return err
	}

	withDependents, err := c.chartsDir.WithTransitiveDependents(asCharts)
	if err != nil {
		return err
	}
	for _, chartName := range withDependents {
		chartNames.Add(chartName.Name())
	}
	return nil
}

func matchesToReleaseFilter(matches globalFileMatches) terra.ReleaseFilter {
	// match all releases
	if matches.includeAllReleases {
		return filter.Releases().Any()
	}

	// start by matching no releases
	_filter := filter.Releases().Any().Negate()

	// include all env releases if needed
	if matches.includeEnvReleases {
		_filter = _filter.Or(filter.Releases().DestinationMatches(filter.Destinations().IsEnvironment()))
	}

	// include all cluster releases if needed
	if matches.includeClusterReleases {
		_filter = _filter.Or(filter.Releases().DestinationMatches(filter.Destinations().IsCluster()))
	}

	return _filter
}

// parse changed files into a list of changed files
func parseChangedList(inputFile string) ([]string, error) {
	var updatedFiles []string

	content, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, errors.Errorf("error reading changed list %s: %v", inputFile, err)
	}

	buf := bytes.NewBuffer(content)
	scanner := bufio.NewScanner(buf)
	var lineno int
	for scanner.Scan() {
		lineno++
		if err := scanner.Err(); err != nil {
			return nil, errors.Errorf("error scanning changed list %s (line %d): %v", inputFile, lineno, err)
		}
		entry := strings.TrimSpace(scanner.Text())
		if entry == "" {
			continue
		}
		if path.IsAbs(entry) {
			return nil, errors.Errorf("changed list file %s contains absolute path but all paths should be relative to terra-helmfile root (line %d): %s", inputFile, lineno, entry)
		}
		updatedFiles = append(updatedFiles, entry)
	}

	log.Debug().Msgf("Found %d entries in %s", len(updatedFiles), inputFile)

	return updatedFiles, nil
}

// split a filepath into a list of components and return the component at the given index
// eg. pathIndex("my/file/path", 1) -> "file"
func pathIndex(file string, index int) string {
	components := strings.Split(file, string(os.PathSeparator))
	return components[index]
}

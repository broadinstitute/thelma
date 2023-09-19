// Package filetrigger maps a list of updated files in the terra-helmfile repo to a list of charts that need to be published.
package filetrigger

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

const FlagName = "changed-files-list"

func ChartList(triggerInputFile string, state terra.State) ([]string, error) {
	files, err := ParseTriggers(triggerInputFile)
	if err != nil {
		return nil, err
	}
	releases, err := state.Releases().Filter(ReleaseFilter(files...))
	if err != nil {
		return nil, err
	}

	chartNames := set.NewSet[string]()
	for _, r := range releases {
		chartNames.Add(r.ChartName())
	}

	charts := chartNames.Elements()
	sort.Strings(charts)
	return charts, nil
}

func ReleaseFilter(updatedFiles ...string) terra.ReleaseFilter {
	opts := struct {
		// include all releases that use one of these charts
		chartNames set.Set[string]
		// include releases that belong to an environment
		includeEnvReleases bool
		// include cluster releases (releases that don't belong to an environment)
		includeClusterReleases bool
		// include all releases
		includeAllReleases bool
	}{
		chartNames: set.NewSet[string](),
	}

	for _, updatedFileName := range updatedFiles {
		// remove . and .. components in path
		cleaned := path.Clean(updatedFileName)

		// loop through files and update filter options based on which files are included
		if regexp.MustCompile("^charts/[^/]").MatchString(cleaned) {
			chartName := pathIndex(cleaned, 1)
			opts.chartNames.Add(chartName)

		} else if regexp.MustCompile("^values/(app|cluster)/global").MatchString(cleaned) {
			kind := pathIndex(cleaned, 1)
			if kind == "app" {
				opts.includeEnvReleases = true
			} else {
				opts.includeClusterReleases = true
			}

		} else if regexp.MustCompile("^values/[^/]").MatchString(cleaned) {
			component := pathIndex(cleaned, 2)
			// remove any . extensions like .yaml, .yaml.gotmpl
			chartName := strings.Split(component, ".")[0]
			opts.chartNames.Add(chartName)

		} else if regexp.MustCompile("^helmfile.yaml$").MatchString(cleaned) {
			opts.includeAllReleases = true
		}
	}

	// return a release filter that matches the options
	if opts.includeAllReleases {
		return filter.Releases().Any()
	}

	// start by matching no releases
	_filter := filter.Releases().Any().Negate()

	// include all env releases if needed
	if opts.includeEnvReleases {
		_filter = _filter.Or(filter.Releases().DestinationMatches(filter.Destinations().IsEnvironment()))
	}

	// include all cluster releases if needed
	if opts.includeClusterReleases {
		_filter = _filter.Or(filter.Releases().DestinationMatches(filter.Destinations().IsCluster()))
	}

	// include all releases that use one an updated charts
	if !opts.chartNames.Empty() {
		_filter = _filter.Or(filter.Releases().HasChartName(opts.chartNames.Elements()...))
	}

	return _filter
}

// ParseTriggers takes a list of trigger-formated files, where each trigger file contains
// a newline-separated list of files that have been updated in the terra-helmfile repo,
// and returns a list of all the entries in each file
func ParseTriggers(triggerInputFiles ...string) ([]string, error) {
	var files []string

	for _, triggerInputFile := range triggerInputFiles {
		list, err := scanOneFile(triggerInputFile)
		if err != nil {
			return nil, err
		}
		files = append(files, list...)
	}

	return files, nil
}

func scanOneFile(triggerInputFile string) ([]string, error) {
	var updatedFiles []string

	content, err := os.ReadFile(triggerInputFile)
	if err != nil {
		return nil, fmt.Errorf("error reading trigger file %s: %v", triggerInputFile, err)
	}

	buf := bytes.NewBuffer(content)
	scanner := bufio.NewScanner(buf)
	var lineno int
	for scanner.Scan() {
		lineno++
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error scanning trigger file %s (line %d): %v", triggerInputFile, lineno, err)
		}
		entry := strings.TrimSpace(scanner.Text())
		if entry == "" {
			continue
		}
		if path.IsAbs(entry) {
			return nil, fmt.Errorf("trigger file %s contains absolute path but all paths should be relative to terra-helmfile root (line %d): %s", triggerInputFile, lineno, entry)
		}
		updatedFiles = append(updatedFiles, entry)
	}

	log.Debug().Msgf("Found %d entries in %s", len(updatedFiles), triggerInputFile)

	return updatedFiles, nil
}

// split a filepath into a list of components and return the component at the given index
// eg. pathIndex("my/file/path", 1) -> "file"
func pathIndex(file string, index int) string {
	components := strings.Split(file, string(os.PathSeparator))
	return components[index]
}

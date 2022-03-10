package selector

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/terra/sort"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// allSelector is used in release selectors to mean "match all releases", "all environments", etc.
const allSelector = "ALL"

// selectorSeparator is used in release selectors to supply multiple comma-separated options
const selectorSeparator = ","

// ReleasesFlagName constant used for the --releases flag name (public because the render package creates an alias for this flag)
const ReleasesFlagName = "releases"

type Selector struct {
	flags         []*enumFlag
	filterBuilder *filterBuilder
}

// Selection describes the set of releases that match user-supplied CLI flags
type Selection struct {
	// IsReleaseScoped true if the user supplied the names of releases (like "agora", "cromwell"), false if they supplied "ALL"
	IsReleaseScoped bool
	// Releases is the set of matching releases
	Releases []terra.Release
	// SingleChart true if we're using a single release name
	SingleChart bool
	// AppReleasesOnly true if all matched releases are application releases
	AppReleasesOnly bool
}

func NewSelector() *Selector {
	return &Selector{
		filterBuilder: newFilterBuilder(),
		flags: []*enumFlag{
			newReleasesFlag(),
			newEnvironmentsFlag(),
			newClustersFlag(),
			newDestinationTypesFlag(),
			newDestinationBasesFlag(),
			newEnvironmentTemplatesFlag(),
			newEnvironmentLifecyclesFlag(),
		},
	}
}

// AddFlags adds selector CLI flags to cobra command
func (s *Selector) AddFlags(cobraCommand *cobra.Command) {
	for _, flag := range s.flags {
		flag.addToCobraCommand(cobraCommand)
	}
}

func (s *Selector) GetSelection(state terra.State, pflags *pflag.FlagSet, args []string) (*Selection, error) {
	for _, flag := range s.flags {
		if err := flag.processInput(s.filterBuilder, state, pflags.Changed(flag.flagName), args); err != nil {
			return nil, err
		}
	}

	releaseFilter := s.filterBuilder.combine()
	releases, err := state.Releases().Filter(releaseFilter)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("%d releases matched filter: %s", len(releases), releaseFilter.String())

	sort.Releases(releases)

	return &Selection{
		IsReleaseScoped: s.filterBuilder.isReleaseScoped(),
		Releases:        releases,
		SingleChart:     singleChart(releases),
		AppReleasesOnly: appReleasesOnly(releases),
	}, nil
}

func singleChart(releases []terra.Release) bool {
	s := set.NewStringSet()
	for _, r := range releases {
		s.Add(r.Name())
	}
	return s.Size() == 1
}

func appReleasesOnly(releases []terra.Release) bool {
	for _, r := range releases {
		if !r.IsAppRelease() {
			return false
		}
	}
	return true
}

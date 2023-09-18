package selector

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/charts/filetrigger"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/sort"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// allSelector is used in release selectors to mean "match all releases", "all environments", etc.
const allSelector = "ALL"

// selectorSeparator is used in release selectors to supply multiple comma-separated options
const selectorSeparator = ","

// ReleasesFlagName constant used for the --release flag name (public because the render package creates an alias for this flag)
const ReleasesFlagName = "release"

var flagNames = struct {
	release              string
	exactRelease         string
	environment          string
	cluster              string
	environmentLifecycle string
	environmentTemplate  string
	destinationType      string
	destinationBase      string
	fileTrigger          string
}{
	release:              ReleasesFlagName,
	exactRelease:         "exact-release",
	environment:          "environment",
	cluster:              "cluster",
	environmentLifecycle: "environment-lifecycle",
	environmentTemplate:  "environment-template",
	destinationBase:      "destination-base",
	destinationType:      "destination-type",
	fileTrigger:          filetrigger.FlagName,
}

type Option func(*Options)

type Options struct {
	// IncludeBulkFlags include bulk destination selection flags such as --destination-type, --environment-template, and so on
	IncludeBulkFlags bool
	// RequireDestinationOrExact requires either one of --environment or --cluster to be passed or for --exact-release to be
	// used instead of --release (since --exact-release is globally unique, as enforced by Sherlock, it is enough to find
	// the destination on its own).
	RequireDestinationOrExact bool
}

type Selector struct {
	options       Options
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

func NewSelector(options ...Option) *Selector {
	opts := Options{
		IncludeBulkFlags:          true,
		RequireDestinationOrExact: false,
	}
	for _, option := range options {
		option(&opts)
	}

	flags := []*enumFlag{
		newReleasesFlag(),
		newExactReleasesFlag(),
		newEnvironmentsFlag(),
		newClustersFlag(),
		newFileTriggerFlag(),
	}

	if opts.IncludeBulkFlags {
		flags = append(flags,
			newDestinationTypesFlag(),
			newDestinationBasesFlag(),
			newEnvironmentTemplatesFlag(),
			newEnvironmentLifecyclesFlag(),
		)
	}

	return &Selector{
		options:       opts,
		filterBuilder: newFilterBuilder(),
		flags:         flags,
	}
}

// AddFlags adds selector CLI flags to cobra command
func (s *Selector) AddFlags(cobraCommand *cobra.Command) {
	for _, flag := range s.flags {
		flag.addToCobraCommand(cobraCommand)
	}
}

func (s *Selector) GetSelection(state terra.State, pflags *pflag.FlagSet, args []string) (*Selection, error) {
	if err := s.checkRequiredFlags(pflags); err != nil {
		return nil, err
	}
	if err := checkIncompatibleFlags(pflags); err != nil {
		return nil, err
	}

	for _, flag := range s.flags {
		if err := flag.processInput(s.filterBuilder, state, args, pflags); err != nil {
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

func (s *Selector) checkRequiredFlags(flags *pflag.FlagSet) error {
	if s.options.RequireDestinationOrExact {
		// "If -e isn't provided, and -c isn't provided, and the user hasn't passed just --exact-release instead of --r"
		if !flags.Changed(flagNames.environment) && !flags.Changed(flagNames.cluster) &&
			!(flags.Changed(flagNames.exactRelease) && !flags.Changed(flagNames.release)) {
			return fmt.Errorf("please specify a target environment or cluster with the -e/-c flags, or specify only full Sherlock-style release names with --exact-release")
		}
	}
	return nil
}

func checkIncompatibleFlags(flags *pflag.FlagSet) error {
	unionFlags := []string{flagNames.environment, flagNames.cluster}

	intersectFlags := []string{flagNames.environmentTemplate, flagNames.environmentLifecycle, flagNames.destinationBase, flagNames.environmentTemplate}

	for _, unf := range unionFlags {
		if flags.Changed(unf) {
			for _, inf := range intersectFlags {
				if flags.Changed(inf) {
					return fmt.Errorf("--%s cannot be combined with --%s", unf, inf)
				}
			}
		}
	}

	return nil
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

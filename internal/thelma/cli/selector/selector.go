package selector

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/changedfiles"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
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
	changedFilesList     string
}{
	release:              ReleasesFlagName,
	exactRelease:         "exact-release",
	environment:          "environment",
	cluster:              "cluster",
	environmentLifecycle: "environment-lifecycle",
	environmentTemplate:  "environment-template",
	destinationBase:      "destination-base",
	destinationType:      "destination-type",
	changedFilesList:     changedfiles.FlagName,
}

type Selector struct {
	flags         []*enumFlag
	filterBuilder *filterBuilder
}

func NewSelector() *Selector {
	flags := []*enumFlag{
		newReleasesFlag(),
		newExactReleasesFlag(),
		newEnvironmentsFlag(),
		newClustersFlag(),
	}

	return &Selector{
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

func (s *Selector) GetSelection(state terra.State, pflags *pflag.FlagSet, args []string) ([]terra.Release, error) {
	if err := s.checkRequiredFlags(pflags); err != nil {
		return nil, err
	}
	if err := checkIncompatibleEnumFlags(pflags); err != nil {
		return nil, err
	}

	for _, flag := range s.flags {
		if err := flag.processInput(s.filterBuilder, state, args, pflags); err != nil {
			return nil, err
		}
	}

	releaseFilter := s.filterBuilder.combine()
	releases, err := applyFilter(state, releaseFilter)
	if err != nil {
		return nil, err
	}

	return releases, nil
}

func (s *Selector) checkRequiredFlags(flags *pflag.FlagSet) error {
	// "If -e isn't provided, and -c isn't provided, and the user hasn't passed just --exact-release instead of --r"
	if !flags.Changed(flagNames.environment) && !flags.Changed(flagNames.cluster) &&
		!(flags.Changed(flagNames.exactRelease) && !flags.Changed(flagNames.release)) {
		return errors.Errorf("please specify a target environment or cluster with the -e/-c flags, or specify only full Sherlock-style release names with --exact-release")
	}
	return nil
}

func checkIncompatibleEnumFlags(flags *pflag.FlagSet) error {
	unionFlags := []string{flagNames.environment, flagNames.cluster}

	intersectFlags := []string{flagNames.environmentTemplate, flagNames.environmentLifecycle, flagNames.destinationBase, flagNames.destinationType}

	for _, unf := range unionFlags {
		if flags.Changed(unf) {
			for _, inf := range intersectFlags {
				if flags.Changed(inf) {
					return errors.Errorf("--%s cannot be combined with --%s", unf, inf)
				}
			}
		}
	}

	return nil
}

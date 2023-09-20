package selector

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RenderSelector struct {
	enumFlags        []*enumFlag
	changedFilesFlag *changedFilesListFlag
	filterBuilder    *filterBuilder
}

// RenderSelection describes the set of releases that match user-supplied CLI flags for a render selector
type RenderSelection struct {
	// IsReleaseScoped true if the user supplied the names of releases (like "agora", "cromwell"), false if they supplied "ALL"
	IsReleaseScoped bool
	// Releases is the set of matching releases
	Releases []terra.Release
	// SingleChart true if we're using a single release name
	SingleChart bool
	// AppReleasesOnly true if all matched releases are application releases
	AppReleasesOnly bool
}

func NewRenderSelector() *RenderSelector {
	enumFlags := []*enumFlag{
		newReleasesFlag(),
		newExactReleasesFlag(),
		newEnvironmentsFlag(),
		newClustersFlag(),
		newDestinationTypesFlag(),
		newDestinationBasesFlag(),
		newEnvironmentTemplatesFlag(),
		newEnvironmentLifecyclesFlag(),
	}

	return &RenderSelector{
		enumFlags:        enumFlags,
		changedFilesFlag: newChangedFilesList(),
		filterBuilder:    newFilterBuilder(),
	}
}

// AddFlags adds selector CLI flags to cobra command
func (s *RenderSelector) AddFlags(cobraCommand *cobra.Command) {
	for _, flag := range s.enumFlags {
		flag.addToCobraCommand(cobraCommand)
	}
	s.changedFilesFlag.addToCobraCommand(cobraCommand)
}

func (s *RenderSelector) GetSelection(state terra.State, chartsDir source.ChartsDir, pflags *pflag.FlagSet, args []string) (*RenderSelection, error) {
	if err := s.checkRequiredFlags(pflags); err != nil {
		return nil, err
	}
	if err := checkIncompatibleEnumFlags(pflags); err != nil {
		return nil, err
	}

	for _, flag := range s.enumFlags {
		if err := flag.processInput(s.filterBuilder, state, args, pflags); err != nil {
			return nil, err
		}
	}
	if err := s.changedFilesFlag.processInput(s.filterBuilder, state, chartsDir, args, pflags); err != nil {
		return nil, err
	}

	releaseFilter := s.filterBuilder.combine()
	releases, err := applyFilter(state, releaseFilter)
	if err != nil {
		return nil, err
	}

	return &RenderSelection{
		Releases:        releases,
		IsReleaseScoped: s.filterBuilder.isReleaseScoped(),
		SingleChart:     singleChart(releases),
		AppReleasesOnly: appReleasesOnly(releases),
	}, nil
}

func (s *RenderSelector) checkRequiredFlags(flags *pflag.FlagSet) error {
	// no required flags yet
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

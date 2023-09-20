package selector

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/sort"
	"github.com/rs/zerolog/log"
)

func newFilterBuilder() *filterBuilder {
	return &filterBuilder{}
}

// filterBuilder is for aggregating a number of complex filters into a single release filter
type filterBuilder struct {
	// holds destination filters (intersected). Used for --destination-type flags and friends
	destinationFilters []terra.DestinationFilter
	// holds release filters (intersected). Used for -r / --release flag
	releaseFilters []terra.ReleaseFilter
	// holds environment filters (intersected). Used for --environment-lifecycle & friends
	environmentFilters []terra.EnvironmentFilter
	// holds union filters (any can be matched). Used for -e / -c flags, which are additive
	destinationIncludes []terra.DestinationFilter
}

// combine aggregates all registered filters into a single terra.ReleaseFilter
func (f *filterBuilder) combine() terra.ReleaseFilter {
	var destFilters []terra.DestinationFilter

	if len(f.destinationIncludes) > 0 {
		// union destinationIncludes and add to the destination filter list
		destFilters = append(destFilters, filter.Destinations().Or(f.destinationIncludes...))
	} else {
		// add intersection destinations
		destFilters = append(destFilters, f.destinationFilters...)
		// convert environment filters into a destination filter that permits environments matching the filter
		if len(f.environmentFilters) > 0 {
			// intersect all env filters
			envFilter := filter.Environments().And(f.environmentFilters...)
			// convert to a destination filter
			destFilter := filter.Destinations().IsEnvironmentMatching(envFilter)
			// permit clusters as well -- --destination-types can be used to restrict typ if so desired
			destFilter = destFilter.Or(filter.Destinations().IsCluster())

			destFilters = append(destFilters, destFilter)
		}
	}

	var releaseFilters []terra.ReleaseFilter
	releaseFilters = append(releaseFilters, f.releaseFilters...)

	// aggregate destination filters with And, then convert the aggregated filter into a release filter
	if len(destFilters) > 0 {
		// intersect all destination filters
		combinedDestFilter := filter.Destinations().And(destFilters...)
		releaseFilters = append(releaseFilters, filter.Releases().DestinationMatches(combinedDestFilter))
	}

	// finally, intersect all release filters
	return filter.Releases().And(releaseFilters...)
}

func (f *filterBuilder) isReleaseScoped() bool {
	return len(f.releaseFilters) > 0
}

func (f *filterBuilder) addEnvironmentFilter(filter terra.EnvironmentFilter) {
	f.environmentFilters = append(f.environmentFilters, filter)
}

func (f *filterBuilder) addDestinationFilter(filter terra.DestinationFilter) {
	f.destinationFilters = append(f.destinationFilters, filter)
}

func (f *filterBuilder) addReleaseFilter(filter terra.ReleaseFilter) {
	f.releaseFilters = append(f.releaseFilters, filter)
}

func (f *filterBuilder) addDestinationInclude(filter terra.DestinationFilter) {
	f.destinationIncludes = append(f.destinationIncludes, filter)
}

func applyFilter(state terra.State, filter terra.ReleaseFilter) ([]terra.Release, error) {
	releases, err := state.Releases().Filter(filter)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("%d releases matched filter: %s", len(releases), filter.String())
	sort.Releases(releases)

	return releases, nil
}

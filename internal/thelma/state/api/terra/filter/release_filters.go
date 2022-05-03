package filter

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// ReleaseFilters provides helper functions for building terra.ReleaseFilter instances
type ReleaseFilters interface {
	// Any returns a filter that matches any release
	Any() terra.ReleaseFilter
	// HasName returns a filter that matches releases with the given name(s)
	HasName(releaseName ...string) terra.ReleaseFilter
	// HasDestinationName returns a filter that matches releases with the given destination
	HasDestinationName(destinationName string) terra.ReleaseFilter
	// DestinationMatches returns a filter that matches releases with matching destinations
	DestinationMatches(filter terra.DestinationFilter) terra.ReleaseFilter
	// BelongsToEnvironment returns a filter that matches releases in the given environment
	BelongsToEnvironment(env terra.Environment) terra.ReleaseFilter
	// Or returns a filter that matches environments that match _any_ of the given filters
	Or(filters ...terra.ReleaseFilter) terra.ReleaseFilter
	//And returns a filter that matches environments that match _all_ of the given filters
	And(filters ...terra.ReleaseFilter) terra.ReleaseFilter
}

// Releases is for building terra.ReleaseFilter instances
func Releases() ReleaseFilters {
	return releaseFilters{}
}

// implements the ReleaseFilters interface
type releaseFilters struct{}

func (r releaseFilters) Any() terra.ReleaseFilter {
	return releaseFilter{
		string: anyString,
		matcher: func(_ terra.Release) bool {
			return true
		},
	}
}

func (r releaseFilters) HasName(releaseNames ...string) terra.ReleaseFilter {
	return releaseFilter{
		string: fmt.Sprintf("hasName(%s)", join(quote(releaseNames)...)),
		matcher: func(r terra.Release) bool {
			for _, name := range releaseNames {
				if r.Name() == name {
					return true
				}
			}
			return false
		},
	}
}

func (r releaseFilters) HasDestinationName(destinationName string) terra.ReleaseFilter {
	return releaseFilter{
		string: fmt.Sprintf("hasDestinationName(%q)", destinationName),
		matcher: func(r terra.Release) bool {
			return r.Destination().Name() == destinationName
		},
	}
}

func (r releaseFilters) DestinationMatches(filter terra.DestinationFilter) terra.ReleaseFilter {
	return releaseFilter{
		string: fmt.Sprintf("destinationMatches(%s)", filter.String()),
		matcher: func(r terra.Release) bool {
			return filter.Matches(r.Destination())
		},
	}
}

func (r releaseFilters) BelongsToEnvironment(env terra.Environment) terra.ReleaseFilter {
	return r.DestinationMatches(Destinations().IsEnvironment().And(Destinations().HasName(env.Name())))
}

//
// TODO [generics] Or and And functions are duplicated across all filter types, fix when generics are available

func (r releaseFilters) Or(filters ...terra.ReleaseFilter) terra.ReleaseFilter {
	if len(filters) == 0 {
		return r.Any()
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return releaseFilter{
		string: fmt.Sprintf(orFormat, join(releaseFilterStrings(filters)...)),
		matcher: func(release terra.Release) bool {
			for _, f := range filters {
				if f.Matches(release) {
					return true
				}
			}
			return false
		},
	}
}

func (r releaseFilters) And(filters ...terra.ReleaseFilter) terra.ReleaseFilter {
	if len(filters) == 0 {
		return r.Any()
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return releaseFilter{
		string: fmt.Sprintf(andFormat, join(releaseFilterStrings(filters)...)),
		matcher: func(release terra.Release) bool {
			for _, f := range filters {
				if !f.Matches(release) {
					return false
				}
			}
			return true
		},
	}
}

func releaseFilterStrings(filters []terra.ReleaseFilter) []string {
	var filterStrings []string
	for _, f := range filters {
		filterStrings = append(filterStrings, f.String())
	}
	return filterStrings
}

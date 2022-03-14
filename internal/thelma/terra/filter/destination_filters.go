package filter

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
)

func Destinations() DestinationFilters {
	return destinationFilters{}
}

// DestinationFilters provides helper functions for building terra.DestinationFilter instances
type DestinationFilters interface {
	// Any returns a filter that matches any destination
	Any() terra.DestinationFilter
	// HasName returns a filter that matches destinations with the given name(s)
	HasName(names ...string) terra.DestinationFilter
	// HasBase returns a filter that matches destinations with the given configuration base(s)
	HasBase(bases ...string) terra.DestinationFilter
	// OfType returns a filter that matches destinations of the given type
	OfType(destinationType terra.DestinationType) terra.DestinationFilter
	// OfTypeName returns a filter that matches the destination type name
	OfTypeName(typeNames ...string) terra.DestinationFilter
	// IsEnvironmentMatching returns a filter that matches destinations that are environments that match the given terra.EnvironmentFilter
	IsEnvironmentMatching(filter terra.EnvironmentFilter) terra.DestinationFilter
	// IsCluster returns a filter that matches clusters
	IsCluster() terra.DestinationFilter
	// IsEnvironment returns a filter that matches environments
	IsEnvironment() terra.DestinationFilter
	// Or returns a filter that matches destinations that match _any_ of the given filters
	Or(filters ...terra.DestinationFilter) terra.DestinationFilter
	//And returns a filter that matches destinations that match _all_ of the given filters
	And(filters ...terra.DestinationFilter) terra.DestinationFilter
}

// implements the DestinationFilters interface
type destinationFilters struct{}

func (d destinationFilters) Any() terra.DestinationFilter {
	return destinationFilter{
		string: anyString,
		matcher: func(_ terra.Destination) bool {
			return true
		},
	}
}

func (d destinationFilters) HasName(destinationNames ...string) terra.DestinationFilter {
	return destinationFilter{
		string: fmt.Sprintf("hasName(%s)", join(quote(destinationNames)...)),
		matcher: func(destination terra.Destination) bool {
			for _, name := range destinationNames {
				if destination.Name() == name {
					return true
				}
			}
			return false
		},
	}
}

func (d destinationFilters) HasBase(destinationBases ...string) terra.DestinationFilter {
	return destinationFilter{
		string: fmt.Sprintf("hasBase(%s)", join(quote(destinationBases)...)),
		matcher: func(destination terra.Destination) bool {
			for _, base := range destinationBases {
				if destination.Base() == base {
					return true
				}
			}
			return false
		},
	}
}

func (d destinationFilters) OfType(destinationType terra.DestinationType) terra.DestinationFilter {
	return destinationFilter{
		string: fmt.Sprintf("ofType(%s)", destinationType.String()),
		matcher: func(d terra.Destination) bool {
			return d.Type() == destinationType
		},
	}
}

func (d destinationFilters) OfTypeName(typeNames ...string) terra.DestinationFilter {
	return destinationFilter{
		string: fmt.Sprintf("ofTypeName(%s)", join(quote(typeNames)...)),
		matcher: func(d terra.Destination) bool {
			for _, typeName := range typeNames {
				if d.Type().String() == typeName {
					return true
				}
			}
			return false
		},
	}
}

func (d destinationFilters) IsEnvironmentMatching(filter terra.EnvironmentFilter) terra.DestinationFilter {
	return destinationFilter{
		string: fmt.Sprintf("isEnvironmentMatching(%s)", filter.String()),
		matcher: func(destination terra.Destination) bool {
			if !destination.IsEnvironment() {
				return false
			}
			env := destination.(terra.Environment)
			return filter.Matches(env)
		},
	}
}

func (d destinationFilters) IsCluster() terra.DestinationFilter {
	return d.OfType(terra.ClusterDestination)
}

func (d destinationFilters) IsEnvironment() terra.DestinationFilter {
	return d.OfType(terra.EnvironmentDestination)
}

//
// TODO [generics] Or and And functions are duplicated across all filter types, fix when generics are available

func (d destinationFilters) Or(filters ...terra.DestinationFilter) terra.DestinationFilter {
	if len(filters) == 0 {
		return d.Any()
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return destinationFilter{
		string: fmt.Sprintf(orFormat, join(destinationFilterStrings(filters)...)),
		matcher: func(destination terra.Destination) bool {
			for _, f := range filters {
				if f.Matches(destination) {
					return true
				}
			}
			return false
		},
	}
}

func (d destinationFilters) And(filters ...terra.DestinationFilter) terra.DestinationFilter {
	if len(filters) == 0 {
		return d.Any()
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return destinationFilter{
		string: fmt.Sprintf(andFormat, join(destinationFilterStrings(filters)...)),
		matcher: func(destination terra.Destination) bool {
			for _, f := range filters {
				if !f.Matches(destination) {
					return false
				}
			}
			return true
		},
	}
}

func destinationFilterStrings(filters []terra.DestinationFilter) []string {
	var filterStrings []string
	for _, f := range filters {
		filterStrings = append(filterStrings, f.String())
	}
	return filterStrings
}

package filter

import "github.com/broadinstitute/thelma/internal/thelma/terra"

func Destinations() DestinationFilters {
	return destinationFilters{}
}

// DestinationFilters provides helper functions for building terra.DestinationFilter instances
type DestinationFilters interface {
	// Any returns a filter that matches any destination
	Any() terra.DestinationFilter
	// OfType returns a filter that matches destinations of the given type
	OfType(destinationType terra.DestinationType) terra.DestinationFilter
}

// implements the DestinationFilters interface
type destinationFilters struct{}

func (d destinationFilters) Any() terra.DestinationFilter {
	return destinationFilter{
		matcher: func(_ terra.Destination) bool {
			return true
		},
	}
}

func (d destinationFilters) OfType(destinationType terra.DestinationType) terra.DestinationFilter {
	return destinationFilter{
		matcher: func(d terra.Destination) bool {
			return d.Type() == destinationType
		},
	}
}

package filter

import "github.com/broadinstitute/thelma/internal/thelma/terra"

// TODO replace separate environment, release, and destination filter implementations with generics once they're available

// implements the terra.DestinationFilter interface
type destinationFilter struct {
	matcher func(terra.Destination) bool
}

func (f destinationFilter) Matches(destination terra.Destination) bool {
	return f.matcher(destination)
}

func (f destinationFilter) And(other terra.DestinationFilter) terra.DestinationFilter {
	return destinationFilter{
		matcher: func(destination terra.Destination) bool {
			return f.Matches(destination) && other.Matches(destination)
		},
	}
}

func (f destinationFilter) Or(other terra.DestinationFilter) terra.DestinationFilter {
	return destinationFilter{
		matcher: func(destination terra.Destination) bool {
			return f.Matches(destination) || other.Matches(destination)
		},
	}
}

package filter

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
)

// TODO replace separate environment, release, and destination filter implementations with generics once they're available

// implements the terra.DestinationFilter interface
type destinationFilter struct {
	string  string
	matcher func(terra.Destination) bool
}

func (f destinationFilter) Matches(destination terra.Destination) bool {
	return f.matcher(destination)
}

func (f destinationFilter) String() string {
	return f.string
}

func (f destinationFilter) And(other terra.DestinationFilter) terra.DestinationFilter {
	return destinationFilter{
		string: fmt.Sprintf(andFormat, join(f.String(), other.String())),
		matcher: func(destination terra.Destination) bool {
			return f.Matches(destination) && other.Matches(destination)
		},
	}
}

func (f destinationFilter) Or(other terra.DestinationFilter) terra.DestinationFilter {
	return destinationFilter{
		string: fmt.Sprintf(orFormat, join(f.String(), other.String())),
		matcher: func(destination terra.Destination) bool {
			return f.Matches(destination) || other.Matches(destination)
		},
	}
}

func (f destinationFilter) Filter(destinations []terra.Destination) []terra.Destination {
	var result []terra.Destination
	for _, destination := range destinations {
		if f.Matches(destination) {
			result = append(result, destination)
		}
	}
	return result
}

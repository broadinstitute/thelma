package filter

import "github.com/broadinstitute/thelma/internal/thelma/terra"

// ReleaseFilters provides helper functions for building terra.ReleaseFilter instances
type ReleaseFilters interface {
	// Any returns a filter that matches any release
	Any() terra.ReleaseFilter
	// HasName returns a filter that matches releases with the given name
	HasName(releaseName string) terra.ReleaseFilter
	// HasDestinationName returns a filter that matches releases with the given destination
	HasDestinationName(destinationName string) terra.ReleaseFilter
}

// Releases is for building terra.ReleaseFilter instances
func Releases() ReleaseFilters {
	return releaseFilters{}
}

// implements the ReleaseFilters interface
type releaseFilters struct{}

func (r releaseFilters) Any() terra.ReleaseFilter {
	return releaseFilter{
		matcher: func(_ terra.Release) bool {
			return true
		},
	}
}

func (r releaseFilters) HasName(releaseName string) terra.ReleaseFilter {
	return releaseFilter{
		matcher: func(r terra.Release) bool {
			return r.Name() == releaseName
		},
	}
}

func (r releaseFilters) HasDestinationName(destinationName string) terra.ReleaseFilter {
	return releaseFilter{
		matcher: func(r terra.Release) bool {
			return r.Destination().Name() == destinationName
		},
	}
}

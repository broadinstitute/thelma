package filter

import "github.com/broadinstitute/thelma/internal/thelma/terra"

// TODO replace separate environment, release, and destination filter implementations with generics once they're available

// implements the terra.ReleaseFilter interface
type releaseFilter struct {
	matcher func(terra.Release) bool
}

func (f releaseFilter) Matches(release terra.Release) bool {
	return f.matcher(release)
}

func (f releaseFilter) And(other terra.ReleaseFilter) terra.ReleaseFilter {
	return releaseFilter{
		matcher: func(release terra.Release) bool {
			return f.Matches(release) && other.Matches(release)
		},
	}
}

func (f releaseFilter) Or(other terra.ReleaseFilter) terra.ReleaseFilter {
	return releaseFilter{
		matcher: func(release terra.Release) bool {
			return f.Matches(release) || other.Matches(release)
		},
	}
}

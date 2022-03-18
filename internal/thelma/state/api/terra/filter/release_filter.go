package filter

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// TODO replace separate environment, release, and destination filter implementations with generics once they're available

// implements the terra.ReleaseFilter interface
type releaseFilter struct {
	string  string
	matcher func(terra.Release) bool
}

func (f releaseFilter) String() string {
	return f.string
}

func (f releaseFilter) Matches(release terra.Release) bool {
	return f.matcher(release)
}

func (f releaseFilter) And(other terra.ReleaseFilter) terra.ReleaseFilter {
	return releaseFilter{
		string: fmt.Sprintf(andFormat, join(f.String(), other.String())),
		matcher: func(release terra.Release) bool {
			return f.Matches(release) && other.Matches(release)
		},
	}
}

func (f releaseFilter) Or(other terra.ReleaseFilter) terra.ReleaseFilter {
	return releaseFilter{
		string: fmt.Sprintf(orFormat, join(f.String(), other.String())),
		matcher: func(release terra.Release) bool {
			return f.Matches(release) || other.Matches(release)
		},
	}
}

func (f releaseFilter) Filter(releases []terra.Release) []terra.Release {
	var result []terra.Release
	for _, release := range releases {
		if f.Matches(release) {
			result = append(result, release)
		}
	}
	return result
}

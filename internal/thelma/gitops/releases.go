package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/terra/compare"
	"sort"
)

// implements the terra.Releases interface
type releases struct {
	state *gitops
}

func newReleases(g *gitops) terra.Releases {
	return &releases{
		state: g,
	}
}

func (r *releases) All() ([]terra.Release, error) {
	return r.Filter(terra.AnyRelease())
}

func (r *releases) Filter(filter terra.ReleaseFilter) ([]terra.Release, error) {
	var result []terra.Release

	allDestinations, err := r.state.Destinations().All()
	if err != nil {
		return nil, err
	}

	for _, _destination := range allDestinations {
		for _, _release := range _destination.Releases() {
			if filter.Matches(_release) {
				result = append(result, _release)
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return compare.Releases(result[i], result[j]) < 0
	})

	return result, nil
}

package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
)

// implements the terra.Releases interface
type releases struct {
	state *state
}

func newReleases(g *state) terra.Releases {
	return &releases{
		state: g,
	}
}

func (r *releases) All() ([]terra.Release, error) {
	var result []terra.Release

	allDestinations, err := r.state.Destinations().All()
	if err != nil {
		return nil, err
	}

	for _, _destination := range allDestinations {
		result = append(result, _destination.Releases()...)
	}

	return result, nil
}

func (r *releases) Filter(filter terra.ReleaseFilter) ([]terra.Release, error) {
	all, err := r.All()
	if err != nil {
		return nil, err
	}

	return filter.Filter(all), nil
}

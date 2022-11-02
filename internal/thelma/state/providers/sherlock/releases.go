package sherlock

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

type releases struct {
	state *state
}

func newReleasesView(s *state) terra.Releases {
	return &releases{
		state: s,
	}
}

func (r *releases) All() ([]terra.Release, error) {
	var result []terra.Release

	allDestinations, err := r.state.Destinations().All()
	if err != nil {
		return nil, err
	}

	for _, destination := range allDestinations {
		result = append(result, destination.Releases()...)
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

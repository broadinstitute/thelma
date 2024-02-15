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

	for _, r := range r.state.releases {
		result = append(result, r)
	}

	return result, nil
}

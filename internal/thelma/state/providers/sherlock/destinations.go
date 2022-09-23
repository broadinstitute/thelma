package sherlock

import (
	"sort"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/compare"
)

type destinations struct {
	state *state
}

func newDestinations(s *state) terra.Destinations {
	return &destinations{
		state: s,
	}
}

func (d *destinations) All() ([]terra.Destination, error) {
	var result []terra.Destination

	for _, env := range d.state.environments {
		result = append(result, env)
	}

	for _, cluster := range d.state.clusters {
		result = append(result, cluster)
	}

	sort.Slice(result, func(i, j int) bool {
		return compare.Destinations(result[i], result[j]) < 0
	})

	return result, nil
}

func (d *destinations) Filter(filter terra.DestinationFilter) ([]terra.Destination, error) {
	all, err := d.All()
	if err != nil {
		return nil, err
	}

	return filter.Filter(all), nil
}

func (d *destinations) Get(name string) (terra.Destination, error) {
	if destination, exists := d.state.clusters[name]; exists {
		return destination, nil
	}

	if destination, exists := d.state.environments[name]; exists {
		return destination, nil
	}

	return nil, nil
}

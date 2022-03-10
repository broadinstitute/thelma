package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/terra/compare"
	"sort"
)

// implements the terra.Destinations interface
type destinations struct {
	state *state
}

func newDestinations(g *state) terra.Destinations {
	return &destinations{
		state: g,
	}
}

func (d *destinations) All() ([]terra.Destination, error) {
	var result []terra.Destination
	for _, env := range d.state.environments {
		result = append(result, env)
	}
	for _, _cluster := range d.state.clusters {
		result = append(result, _cluster)
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
	if _destination, exists := d.state.clusters[name]; exists {
		return _destination, nil
	}

	if _destination, exists := d.state.environments[name]; exists {
		return _destination, nil
	}

	return nil, nil
}

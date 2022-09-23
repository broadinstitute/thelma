package sherlock

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

type clusters struct {
	state *state
}

func newClusters(s *state) terra.Clusters {
	return &clusters{
		state: s,
	}
}

func (c *clusters) All() ([]terra.Cluster, error) {
	var result []terra.Cluster
	for _, cluster := range c.state.clusters {
		result = append(result, cluster)
	}

	return result, nil
}

func (c *clusters) Get(name string) (terra.Cluster, error) {
	cluster, exists := c.state.clusters[name]
	if !exists {
		return nil, nil
	}

	return cluster, nil
}

func (c *clusters) Exists(name string) (bool, error) {
	_, exists := c.state.clusters[name]
	return exists, nil
}

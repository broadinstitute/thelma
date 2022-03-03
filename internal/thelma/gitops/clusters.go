package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
)

// implements the terra.Clusters interface
type clusters struct {
	state *state
}

func newClusters(g *state) terra.Clusters {
	return &clusters{
		state: g,
	}
}

func (c *clusters) All() ([]terra.Cluster, error) {
	var result []terra.Cluster
	for _, _cluster := range c.state.clusters {
		result = append(result, _cluster)
	}
	return result, nil
}

func (c *clusters) Get(name string) (terra.Cluster, error) {
	return c.state.clusters[name], nil
}

func (c *clusters) Exists(name string) (bool, error) {
	_, exists := c.state.clusters[name]
	return exists, nil
}

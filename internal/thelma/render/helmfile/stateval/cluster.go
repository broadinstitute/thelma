package stateval

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Cluster -- information about the cluster the release is being deployed to
type Cluster struct {
	// Name of the cluster this release is being deployed to (omitted if app release)
	Name string `yaml:"Name,omitempty"`
}

func forCluster(cluster terra.Cluster) Cluster {
	return Cluster{
		Name: cluster.Name(),
	}
}

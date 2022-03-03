package stateval

import "github.com/broadinstitute/thelma/internal/thelma/terra"

// Cluster -- information about the cluster the release is being deployed to
type Cluster struct {
	// Name of the cluster this release is being deployed to (omitted if app release)
	Name string `yaml:"Name,omitempty"`
}

type ClusterValues struct {
	destination Destination
	cluster     Cluster
}

func forCluster(cluster terra.Cluster) Cluster {
	return Cluster{
		Name: cluster.Name(),
	}
}

package gitops

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// implements the terra.ClusterRelease interface
type clusterRelease struct {
	release
}

func (r *clusterRelease) Cluster() terra.Cluster {
	return r.destination.(terra.Cluster)
}

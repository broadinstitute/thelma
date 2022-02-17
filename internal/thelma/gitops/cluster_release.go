package gitops

import "github.com/broadinstitute/thelma/internal/thelma/terra"

type clusterRelease struct {
	release
}

func (r *clusterRelease) Cluster() terra.Cluster {
	return r.destination.(terra.Cluster)
}

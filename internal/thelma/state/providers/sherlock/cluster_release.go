package sherlock

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

type clusterRelease struct {
	release
}

func (r *clusterRelease) Cluster() terra.Cluster {
	return r.cluster
}

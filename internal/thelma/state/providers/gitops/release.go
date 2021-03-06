package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type release struct {
	name                string
	enabled             bool
	releaseType         terra.ReleaseType
	chartVersion        string
	chartName           string
	repo                string
	namespace           string
	cluster             terra.Cluster
	destination         terra.Destination
	terraHelmfileRef    string
	firecloudDevelopRef string
}

func (r *release) Name() string {
	return r.name
}

func (r *release) Type() terra.ReleaseType {
	return r.releaseType
}

func (r *release) IsAppRelease() bool {
	return r.Type() == terra.AppReleaseType
}

func (r *release) IsClusterRelease() bool {
	return r.Type() == terra.ClusterReleaseType
}

func (r *release) ChartName() string {
	return r.chartName
}

func (r *release) ChartVersion() string {
	return r.chartVersion
}

func (r *release) Repo() string {
	return r.repo
}

func (r *release) Namespace() string {
	return r.namespace
}

func (r *release) Cluster() terra.Cluster {
	return r.cluster
}

func (r *release) ClusterName() string {
	return r.cluster.Name()
}

func (r *release) ClusterAddress() string {
	return r.cluster.Address()
}

func (r *release) Destination() terra.Destination {
	return r.destination
}

func (r *release) TerraHelmfileRef() string {
	return r.terraHelmfileRef
}

func (r *release) FirecloudDevelopRef() string {
	return r.firecloudDevelopRef
}

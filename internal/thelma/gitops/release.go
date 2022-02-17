package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
)

type release struct {
	name           string
	releaseType    terra.ReleaseType
	chartVersion   string
	chartName      string
	repo           string
	namespace      string
	clusterName    string
	clusterAddress string
	destination    terra.Destination
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

func (r *release) ClusterName() string {
	return r.clusterName
}

func (r *release) ClusterAddress() string {
	return r.clusterAddress
}

func (r *release) Destination() terra.Destination {
	return r.destination
}

package terra

// Release represents a deployed instance of a Helm chart, running in a Kubernetes cluster. The term comes from Helm.
type Release interface {
	Named
	FullName() string
	Type() ReleaseType
	IsAppRelease() bool
	IsClusterRelease() bool
	ChartVersion() string
	ChartName() string
	Repo() string
	Namespace() string
	Cluster() Cluster
	ClusterName() string
	ClusterAddress() string
	Destination() Destination
	TerraHelmfileRef() string
	FirecloudDevelopRef() string
}

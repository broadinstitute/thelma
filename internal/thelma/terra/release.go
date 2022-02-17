package terra

type Release interface {
	Name() string
	Type() ReleaseType
	IsAppRelease() bool
	IsClusterRelease() bool
	ChartVersion() string
	ChartName() string
	Repo() string
	Namespace() string
	ClusterName() string
	ClusterAddress() string
	Destination() Destination
}

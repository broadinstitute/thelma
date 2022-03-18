package terra

type ClusterRelease interface {
	Cluster() Cluster
	Release
}

package terra

// Clusters is an interface for querying clusters
type Clusters interface {
	// All returns a list of all clusters
	All() ([]Cluster, error)
	// Get returns the cluster with the given name, or an error if no such cluster exists
	Get(name string) (Cluster, error)
	// Exists returns true if a cluster by the given name exists
	Exists(name string) (bool, error)
}

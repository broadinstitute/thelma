package terra

// State is an interface for querying the state of Terra infrastructure.
type State interface {
	Destinations() Destinations
	Environments() Environments
	Clusters() Clusters
	Releases() Releases
}

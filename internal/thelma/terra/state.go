package terra

// State is an interface for querying the state of Terra infrastructure.
type State interface {
	// Destinations is an interface for querying terra.Destination instances
	Destinations() Destinations
	// Environments is an interface for querying terra.Environment instances
	Environments() Environments
	// Clusters is an interface for querying terra.Cluster instances
	Clusters() Clusters
	// Releases is an interface for querying terra.Release instances
	Releases() Releases
}

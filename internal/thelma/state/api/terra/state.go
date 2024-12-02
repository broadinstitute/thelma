// Package terra contains interfaces that model Terra's infrastructure, and support querying and updating the state
// of said infrastructure.
package terra

// State is an interface for querying the state of Terra infrastructure.
type State interface {
	// Environments is an interface for querying terra.Environment instances
	Environments() Environments
	// Clusters is an interface for querying terra.Cluster instances
	Clusters() Clusters
	// Releases is an interface for querying terra.Release instances
	Releases() Releases
}

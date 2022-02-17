package terra

// Destination represents where a release is being deployed (environment or cluster)
type Destination interface {
	Type() DestinationType    // Type is the name of the destination type, either "environment" or "cluster" or "environment-template"
	Base() string             // Base is the base of the environment or cluster
	Name() string             // Name is the name of the environment or cluster
	ReleaseType() ReleaseType // ReleaseType returns the types of releases that can be deployed to this destination
	Releases() []Release      // Releases returns the set of releases configured for this destination
	IsCluster() bool          // Returns true if this destination is a cluster
	IsEnvironment() bool      // Returns true if this destination is an environment
}

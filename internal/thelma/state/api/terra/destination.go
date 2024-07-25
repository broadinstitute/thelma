package terra

// Destination is the location where a release is deployed (environment or cluster)
type Destination interface {
	Named                     // Named provides the name of the environment or cluster
	Type() DestinationType    // Type is the name of the destination type, either "environment" or "cluster"
	Base() string             // Base is the base of the environment or cluster
	ReleaseType() ReleaseType // ReleaseType returns the types of releases that can be deployed to this destination
	Releases() []Release      // Releases returns the set of releases configured for this destination
	TerraHelmfileRef() string // TerraHelmfileRef this destination's generator should be pinned to
	IsCluster() bool          // IsCluster Returns true if this destination is a cluster
	IsEnvironment() bool      // IsEnvironment Returns true if this destination is an environment
	RequireSuitable() bool    // RequireSuitable Returns true if this destination requires suitability in order to modify
	RequiredRole() string     // RequiredRole is the role that users must have to mutate this destination. Thelma should pass this value verbatim.
}

package terra

type Environment interface {
	// DefaultCluster Returns the default cluster for this environment.
	DefaultCluster() Cluster
	// Namespace Returns the namespace for this environment. Eg. "terra-dev"
	Namespace() string
	// Lifecycle returns the lifecycle for this environment.
	Lifecycle() Lifecycle
	// Template returns the name of this environment's configuration template, if it has one.
	// Returns the empty string if the environment has no configuration template.
	Template() string
	// IsHybrid DEPRECATED returns true if this is a hybrid environment (connected to a FiaB)
	IsHybrid() bool
	// Fiab DEPRECATED returns the Fiab associated with this hybrid environment (nil if this is not a hybrid environment)
	Fiab() Fiab
	// BaseDomain returns static domain name part for this environment or environment type.
	// E.g. "bee.envs-terra.bio", "dsde-prod.broadinstitute.org"
	BaseDomain() string
	// NamePrefixesDomain returns whether this particular environment's name should come before its BaseDomain when
	// deriving full hostnames/URLs in this environment.
	// E.g. 'true' for dynamic/template environments, 'false' for static
	NamePrefixesDomain() bool
	// BuildNumber returns the current build number for any CI builds actively running against the environment.
	// Returns 0 if no build number has been set.
	BuildNumber() int
	Destination
}

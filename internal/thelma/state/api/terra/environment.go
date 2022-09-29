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
	// BaseDomain returns static domain name part for this environment or environment type.
	// E.g. "bee.envs-terra.bio", "dsde-prod.broadinstitute.org"
	BaseDomain() string
	// NamePrefixesDomain returns whether this particular environment's name should come before its BaseDomain when
	// deriving full hostnames/URLs in this environment.
	// E.g. 'true' for dynamic/template environments, 'false' for static
	NamePrefixesDomain() bool
	// UniqueResourcePrefix (dynamic environments only) unique-to-this-environment 4-character prefix that can be referenced in configuration.
	// Format: [a-z][a-z0-9]{3}
	// Returns empty string for static / template environments
	UniqueResourcePrefix() string
	Destination
}

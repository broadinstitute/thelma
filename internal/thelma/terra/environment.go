package terra

type Environment interface {
	// DefaultCluster Returns the name of the default cluster for this environment. Eg. "terra-qa"
	DefaultCluster() string
	// Namespace Returns the namespace for this environment. Eg. "terra-dev"
	Namespace() string
	// Lifecycle returns the lifecycle for this environment. Eg. "static"
	Lifecycle() Lifecycle
	// Template returns the name of this environment's configuration template, if it has one.
	// Returns the empty string if the environment has no configuration template.
	Template() string
	// HasTemplate returns true if this environment has a configuration template
	HasTemplate() bool
	Destination
}

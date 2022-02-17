package terra

type Environment interface {
	// Returns the name of the default cluster for this environment. Eg. "terra-qa"
	DefaultCluster() string
	// Returns the namespace for this environment. Eg. "terra-dev"
	Namespace() string
	// Lifecycle returns the lifecycle for this environment. Eg. "static"
	Lifecycle() Lifecycle
	Destination
}

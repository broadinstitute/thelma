package terra

// Environments is an interface for querying and updating environments
type Environments interface {
	// All returns a list of all environments
	All() ([]Environment, error)
	// Filter returns a list of environments matching the given filter
	Filter(filter EnvironmentFilter) ([]Environment, error)
	// Get returns the environment with the given name, or nil if no such environment exists
	Get(name string) (Environment, error)
	// Exists returns true if an environment by the given name exists
	Exists(name string) (bool, error)
	// CreateFromTemplate creates a new environment with the given name from the given template.
	// Should panic if the template environment's lifecycle is not "template".
	CreateFromTemplate(name string, template Environment) error
	// CreateHybridFromTemplate creates a new hybrid environment with the given name from the given template.
	CreateHybridFromTemplate(name string, template Environment, fiab Fiab) error
	// PinVersions pins a set of services in an environment to specific versions. This is _additive_. In other words,
	PinVersions(name string, versions map[string]string) error
	// UnpinVersions removes version overrides for the given environment
	UnpinVersions(name string) error
	// Delete deletes the environment with the given name
	Delete(name string) error
}

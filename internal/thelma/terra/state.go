package terra

// State is an interface for querying the state of Terra infrastructure.
type State interface {
	Destinations() Destinations
	Environments() Environments
	Releases() Releases
}

// Destinations is an interface for querying release destinations
type Destinations interface {
	// All returns a list of all destinations
	All() ([]Destination, error)
	// Get returns the destination with the given name, or an error if no such destination exists
	Get(name string) (Destination, error)
}

// Environments is an interface for querying and updating environments
type Environments interface {
	// Environments returns a list of all environments
	All() ([]Environment, error)
	// Get returns the environment with the given name, or nil if no such environment exists
	Get(name string) (Environment, error)
	// Exists returns true if an environment by the given name exists
	Exists(name string) (bool, error)
	// CreateFromTemplate creates a new environment with the given name from the given template.
	// Should panic if the template environment's lifecycle is not "template".
	CreateFromTemplate(name string, template Environment) error
	// PinVersions pins a set of services in an environment to specific versions. This is _additive_. In other words,
	//
	PinVersions(name string, versions map[string]string) error
	// UnpinVersions removes version overrides for the given environment
	UnpinVersions(name string) error
	// Delete deletes the environment with the given name
	Delete(name string) error
}

// Releases is an interface for querying releases
type Releases interface {
	// All returns a list of all releases
	All() ([]Release, error)
	// Filter filters releases
	Filter(filter ReleaseFilter) ([]Release, error)
}

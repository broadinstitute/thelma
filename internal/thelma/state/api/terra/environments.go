package terra

// Environments is an interface for querying and updating environments
type Environments interface {
	// All returns a list of all environments
	All() ([]Environment, error)
	// Get returns the environment with the given name, or nil if no such environment exists
	Get(name string) (Environment, error)
	// Exists returns true if an environment by the given name exists
	Exists(name string) (bool, error)
	// CreateFromTemplate creates a new environment with the given options from the given template.
	// Must panic if the template environment's lifecycle is not "template".
	CreateFromTemplate(template Environment, opts CreateOptions) (string, error)
	// EnableRelease enables a release in an environment
	// TODO this should move to Environment at some point
	EnableRelease(environmentName string, releaseName string) error
	// DisableRelease disables a release in an environment
	// TODO this should move to Environment at some point
	DisableRelease(environmentName string, releaseName string) error
	// PinVersions sets a version override in the given environment
	// TODO this should move to Environment at some point
	PinVersions(environmentName string, versions map[string]VersionOverride) (map[string]VersionOverride, error)
	// UnpinVersions removes version overrides in the given environment
	// TODO this should move to Environment at some point
	UnpinVersions(environmentName string) (map[string]VersionOverride, error)
	// PinEnvironmentToTerraHelmfileRef pins an environment to a specific terra-helmfile ref
	// Note this can be overridden by individual service version overrides
	PinEnvironmentToTerraHelmfileRef(environmentName string, terraHelmfileRef string) error
	// Delete deletes the environment with the given name
	Delete(name string) error
	// SetOffline controls whether an environment is meant to be online or offline.
	SetOffline(name string, offline bool) error
}

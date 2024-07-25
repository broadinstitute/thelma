package stateval

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Destination -- information about where chart release is being deployed (env, cluster)
type Destination struct {
	// Name of the environment or cluster
	Name string `yaml:"Name"`
	// Type -- either "environment" or "cluster"
	Type string `yaml:"Type"`
	// ConfigBase configuration base for this environment or cluster. Eg. "live", "terra"
	ConfigBase string `yaml:"ConfigBase"`
	// ConfigName configuration name for this environment for cluster.
	// (same as Name except for dynamically-created environments)
	ConfigName string `yaml:"ConfigName"`
	// RequireSuitable whether users must be suitable in order to access/deploy to this destination
	RequireSuitable bool `yaml:"RequireSuitable"`
	// RequiredRole indicates the Sherlock role users must have to mutate this destination.
	// Thelma should pass this value verbatim.
	RequiredRole string `yaml:"RequiredRole"`
}

func forDestination(destination terra.Destination) Destination {
	// ConfigName is the same as the environment/cluster name UNLESS this is
	// a dynamically-generated environment that uses a config template... in which
	// case the config name is the name of the template
	configName := destination.Name()
	if destination.IsEnvironment() {
		env := destination.(terra.Environment)
		if env.Lifecycle().IsDynamic() {
			configName = env.Template()
		}
	}

	return Destination{
		Name:            destination.Name(),
		Type:            destination.Type().String(),
		ConfigBase:      destination.Base(),
		ConfigName:      configName,
		RequireSuitable: destination.RequireSuitable(),
		RequiredRole:    destination.RequiredRole(),
	}
}

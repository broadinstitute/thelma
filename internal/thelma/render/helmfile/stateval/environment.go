package stateval

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Environment -- information about the environment the release is being deployed to.
type Environment struct {
	// Name of the environment this release is being deployed to
	Name string `yaml:"Name"`
	// UniqueResourcePrefix for the environment this release is being deployed to
	UniqueResourcePrefix string `yaml:"UniqueResourcePrefix"`
	// EnableJanitor indicates whether the Janitor service should be used for this
	// environment to help reduce cloud costs.
	EnableJanitor bool `yaml:"EnableJanitor"`
}

func forEnvironment(environment terra.Environment) Environment {
	var env Environment
	env.Name = environment.Name()
	env.UniqueResourcePrefix = environment.UniqueResourcePrefix()
	env.EnableJanitor = environment.EnableJanitor()
	return env
}

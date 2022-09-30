package stateval

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Environment -- information about the environment the release is being deployed to.
type Environment struct {
	// Name of the environment this release is being deployed to
	Name string `yaml:"Name"`
	// UniqueResourcePrefix for the environment this release is being deployed to
	UniqueResourcePrefix string `yaml:"UniqueResourcePrefix"`
}

func forEnvironment(environment terra.Environment) Environment {
	var env Environment
	env.Name = environment.Name()
	env.UniqueResourcePrefix = environment.UniqueResourcePrefix()
	return env
}

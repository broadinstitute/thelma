package stateval

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Environment -- information about the environment the release is being deployed to.
type Environment struct {
	// Name of the environment this release is being deployed to
	Name string `yaml:"Name"`
	// BuildNumber Number for a running CI build currently running against this environment (0 if unset)
	BuildNumber int `yaml:"BuildNumber""`
	// DEPRECATED (remove once we are no longer running hybrids bee/fiab envs)
	IsHybrid bool `yaml:"IsHybrid"`
	// DEPRECATED (remove once we are no longer running hybrid bee/fiab envs)
	Fiab struct {
		Name string `yaml:"Name,omitempty"`
		IP   string `yaml:"IP,omitempty"`
	} `yaml:"Fiab,omitempty"`
}

func forEnvironment(environment terra.Environment) Environment {
	var env Environment
	env.Name = environment.Name()
	env.BuildNumber = environment.BuildNumber()
	env.IsHybrid = environment.IsHybrid()
	if environment.IsHybrid() {
		env.Fiab.Name = environment.Fiab().Name()
		env.Fiab.IP = environment.Fiab().IP()
	}
	return env
}

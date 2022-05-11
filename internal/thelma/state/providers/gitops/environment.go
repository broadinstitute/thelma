package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

const envNamespacePrefix = "terra-"

// implements the terra.Environment interface
type environment struct {
	defaultCluster string                 // Name of the default cluster for this environment. eg "terra-dev"
	releases       map[string]*appRelease // Set of releases configured in this environment
	lifecycle      terra.Lifecycle        // Lifecycle for this environment
	template       string                 // Template for this environment, if it has one
	fiab           terra.Fiab             // DEPRECATED fiab associated with this environment, if there is one
	destination
}

// newEnvironment constructs a new Environment
func newEnvironment(name string, base string, defaultCluster string, lifecycle terra.Lifecycle, template string, fiab terra.Fiab, releases map[string]*appRelease) *environment {
	return &environment{
		defaultCluster: defaultCluster,
		releases:       releases,
		lifecycle:      lifecycle,
		template:       template,
		fiab:           fiab,
		destination: destination{
			name:            name,
			base:            base,
			destinationType: terra.EnvironmentDestination,
		},
	}
}

func (e *environment) Releases() []terra.Release {
	var result []terra.Release
	for _, r := range e.releases {
		if r.enabled {
			result = append(result, r)
		}
	}
	return result
}

func (e *environment) DefaultCluster() string {
	return e.defaultCluster
}

func (e *environment) Lifecycle() terra.Lifecycle {
	return e.lifecycle
}

func (e *environment) Template() string {
	return e.template
}

func (e *environment) ReleaseType() terra.ReleaseType {
	return terra.AppReleaseType
}

// Name environment name, eg. "alpha"
func (e *environment) Name() string {
	return e.name
}

// Base environment configuration base, eg. "live"
func (e *environment) Base() string {
	return e.base
}

// Namespace returns the environment's namespace. Eg "terra-dev", "terra-perf", etc.
func (e *environment) Namespace() string {
	return environmentNamespace(e.Name())
}

func (e *environment) IsHybrid() bool {
	return e.fiab != nil
}

func (e *environment) Fiab() terra.Fiab {
	return e.fiab
}

// environmentNamespace return environment namespace for a given environment name
func environmentNamespace(envName string) string {
	return fmt.Sprintf("%s%s", envNamespacePrefix, envName)
}

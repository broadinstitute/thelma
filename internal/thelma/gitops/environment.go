package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
)

const envNamespacePrefix = "terra-"

// Environment represents a Terra environment
type environment struct {
	defaultCluster string                      // Name of the default cluster for this environment. eg "terra-dev"
	releases       map[string]terra.AppRelease // Set of releases configured in this environment
	lifecycle      terra.Lifecycle             // Lifecycle for this environment
	template       string                      // Template for this environment, if it has one
	destination
}

// NewEnvironment constructs a new Environment
func NewEnvironment(name string, base string, defaultCluster string, lifecycle terra.Lifecycle, template string, releases map[string]terra.AppRelease) terra.Environment {
	return &environment{
		defaultCluster: defaultCluster,
		releases:       releases,
		lifecycle:      lifecycle,
		template:       template,
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
		result = append(result, r)
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

func (e *environment) HasTemplate() bool {
	return e.Template() == ""
}

func (e *environment) ReleaseType() terra.ReleaseType {
	return terra.AppReleaseType
}

// Name environment name, eg. "alpha"
func (e *environment) Name() string {
	return e.name
}

// Base environment base, eg. "live"
func (e *environment) Base() string {
	return e.base
}

// Namespace returns the environment's namespace. Eg "terra-dev", "terra-perf", etc.
func (e *environment) Namespace() string {
	return environmentNamespace(e.Name())
}

// environmentNamespace return environment namespace for a given environment name
func environmentNamespace(envName string) string {
	return fmt.Sprintf("%s%s", envNamespacePrefix, envName)
}

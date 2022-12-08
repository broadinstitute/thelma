package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"time"
)

const envNamespacePrefix = "terra-"

// implements the terra.Environment interface
type environment struct {
	defaultCluster       terra.Cluster          // Default cluster for this environment.
	releases             map[string]*appRelease // Set of releases configured in this environment
	lifecycle            terra.Lifecycle        // Lifecycle for this environment
	template             string                 // Template for this environment, if it has one
	baseDomain           string                 // the stable domain part for this environment
	namePrefixesDomain   bool                   // if baseDomain should be prefixed with destination.name
	uniqueResourcePrefix string                 // uniqueResourcePrefix a unique-this-environment prefix that can be referenced in Helm configuration (applies to dynamic environments only)
	destination
}

// newEnvironment constructs a new Environment
func newEnvironment(
	name string,
	base string,
	defaultCluster terra.Cluster,
	lifecycle terra.Lifecycle,
	template string,
	requireSuitable bool,
	baseDomain string,
	namePrefixesDomain bool,
	releases map[string]*appRelease,
	uniqueResourcePrefix string,
) *environment {
	return &environment{
		defaultCluster:       defaultCluster,
		releases:             releases,
		lifecycle:            lifecycle,
		template:             template,
		baseDomain:           baseDomain,
		namePrefixesDomain:   namePrefixesDomain,
		uniqueResourcePrefix: uniqueResourcePrefix,
		destination: destination{
			name:            name,
			base:            base,
			destinationType: terra.EnvironmentDestination,
			requireSuitable: requireSuitable,
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

func (e *environment) DefaultCluster() terra.Cluster {
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

func (e *environment) BaseDomain() string {
	return e.baseDomain
}

func (e *environment) NamePrefixesDomain() bool {
	return e.namePrefixesDomain
}

func (e *environment) UniqueResourcePrefix() string {
	return e.uniqueResourcePrefix
}

func (e *environment) Owner() string {
	// Gitops doesn't track owner
	return ""
}

// Now that the gitops state provider is no longer used except in tests, we're adding some dummy implementations here
// to keep tests compiling until it can be ripped out

func (e *environment) CreatedAt() time.Time {
	return time.Now()
}

func (e *environment) PreventDeletion() bool {
	return true
}

func (e *environment) AutoDelete() terra.AutoDelete {
	return autoDelete{}
}

type autoDelete struct{}

func (a autoDelete) After() time.Time {
	return time.Now()
}

func (a autoDelete) Enabled() bool {
	return false
}

// environmentNamespace return environment namespace for a given environment name
func environmentNamespace(envName string) string {
	return fmt.Sprintf("%s%s", envNamespacePrefix, envName)
}

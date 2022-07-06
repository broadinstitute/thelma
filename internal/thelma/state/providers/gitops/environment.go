package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

const envNamespacePrefix = "terra-"

// implements the terra.Environment interface
type environment struct {
	defaultCluster     terra.Cluster          // Default cluster for this environment.
	releases           map[string]*appRelease // Set of releases configured in this environment
	lifecycle          terra.Lifecycle        // Lifecycle for this environment
	template           string                 // Template for this environment, if it has one
	fiab               terra.Fiab             // DEPRECATED fiab associated with this environment, if there is one
	baseDomain         string                 // the stable domain part for this environment
	namePrefixesDomain bool                   // if baseDomain should be prefixed with destination.name
	buildNumber        int                    // buildNumber number of a CI build running against the environment, if there is one
	destination
}

// newEnvironment constructs a new Environment
func newEnvironment(
	name string,
	base string,
	defaultCluster terra.Cluster,
	lifecycle terra.Lifecycle,
	template string,
	fiab terra.Fiab,
	requireSuitable bool,
	baseDomain string,
	namePrefixesDomain bool,
	releases map[string]*appRelease,
	buildNumber int,
) *environment {
	return &environment{
		defaultCluster:     defaultCluster,
		releases:           releases,
		lifecycle:          lifecycle,
		template:           template,
		fiab:               fiab,
		baseDomain:         baseDomain,
		namePrefixesDomain: namePrefixesDomain,
		buildNumber:        buildNumber,
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

func (e *environment) IsHybrid() bool {
	return e.fiab != nil
}

func (e *environment) Fiab() terra.Fiab {
	return e.fiab
}

func (e *environment) BaseDomain() string {
	return e.baseDomain
}

func (e *environment) NamePrefixesDomain() bool {
	return e.namePrefixesDomain
}

func (e *environment) BuildNumber() int {
	return e.buildNumber
}

// environmentNamespace return environment namespace for a given environment name
func environmentNamespace(envName string) string {
	return fmt.Sprintf("%s%s", envNamespacePrefix, envName)
}

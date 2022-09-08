package sherlock

import (
	"fmt"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

const envNamespacePrefix = "terra-"

type environment struct {
	defaultCluster     terra.Cluster
	releases           map[string]*appRelease
	lifecycle          terra.Lifecycle
	template           string
	baseDomain         string
	namePrefixesDomain bool
	buildNumber        int
	destination
}

func (e *environment) Releases() []terra.Release {
	var result []terra.Release
	for _, release := range e.releases {
		if release.enabled {
			result = append(result, release)
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

func (e *environment) Name() string {
	return e.name
}

func (e *environment) Base() string {
	return e.base
}

func (e *environment) Namespace() string {
	return environmentNamespace(e.Name())
}

func (e *environment) IsHybrid() bool {
	panic("deprecated, implementing for interface compatability")
}

func (e *environment) Fiab() terra.Fiab {
	panic("deprecated, implementing for interface compatability")
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

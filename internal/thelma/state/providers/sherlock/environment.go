package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type environment struct {
	defaultCluster       terra.Cluster
	defaultNamespace     string
	releases             map[string]*appRelease
	lifecycle            terra.Lifecycle
	template             string
	baseDomain           string
	namePrefixesDomain   bool
	uniqueResourcePrefix string
	owner                string
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
	if e.defaultNamespace != "" {
		return e.defaultNamespace
	} else {
		return e.Name()
	}
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
	return e.owner
}

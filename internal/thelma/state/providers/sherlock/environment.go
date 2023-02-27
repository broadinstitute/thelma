package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"time"
)

type environment struct {
	createdAt                   time.Time
	defaultCluster              terra.Cluster
	defaultNamespace            string
	releases                    map[string]*appRelease
	lifecycle                   terra.Lifecycle
	template                    string
	baseDomain                  string
	namePrefixesDomain          bool
	uniqueResourcePrefix        string
	owner                       string
	preventDeletion             bool
	autoDelete                  autoDelete
	offline                     bool
	offlineScheduleBeginEnabled bool
	offlineScheduleBeginTime    time.Time
	offlineScheduleEndEnabled   bool
	offlineScheduleEndTime      time.Time
	offlineScheduleEndWeekends  bool
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

func (e *environment) CreatedAt() time.Time {
	return e.createdAt
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

func (e *environment) PreventDeletion() bool {
	return e.preventDeletion
}

func (e *environment) AutoDelete() terra.AutoDelete {
	return e.autoDelete
}

func (e *environment) Offline() bool {
	return e.offline
}

func (e *environment) OfflineScheduleBeginEnabled() bool {
	return e.offlineScheduleBeginEnabled
}

func (e *environment) OfflineScheduleBeginTime() time.Time {
	return e.offlineScheduleBeginTime
}

func (e *environment) OfflineScheduleEndEnabled() bool {
	return e.offlineScheduleEndEnabled
}

func (e *environment) OfflineScheduleEndTime() time.Time {
	return e.offlineScheduleEndTime
}

func (e *environment) OfflineScheduleEndWeekends() bool {
	return e.offlineScheduleEndWeekends
}

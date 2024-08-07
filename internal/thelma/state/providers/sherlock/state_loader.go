package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

type stateLoader struct {
	sherlock   sherlock.Client
	thelmaHome string
	cached     terra.State
}

func NewStateLoader(thelmaHome string, sherlock sherlock.Client) terra.StateLoader {
	return &stateLoader{
		thelmaHome: thelmaHome,
		sherlock:   sherlock,
	}
}

func (s *stateLoader) Load() (terra.State, error) {
	if s.cached == nil {
		return s.Reload()
	} else {
		return s.cached, nil
	}
}

func (s *stateLoader) Reload() (terra.State, error) {
	// Note that this retry loop is *not* for retrying actual errors. Maybe we'll make that call at some point,
	// but right now if there's an actual error talking to Sherlock or with the returned data, we want to error
	// so we can fix it.
	// This loop is specifically for addressing race conditions as we assemble our state to match Sherlock's.
	// Since we make multiple requests and have to associate those responses, it is possible that Sherlock's
	// state could change between responses, making our data inconsistent. This doesn't represent an error on
	// anyone's part but it is something we need to account for and retry, so here we are.
retry:
	for attempts := 3; attempts > 0; attempts-- {
		stateClusters, err := s.sherlock.Clusters()
		if err != nil {
			return nil, err
		}
		stateEnvironments, err := s.sherlock.Environments()
		if err != nil {
			return nil, err
		}
		stateReleases, err := s.sherlock.Releases()
		if err != nil {
			return nil, err
		}

		_clusters := make(map[string]*cluster)
		for _, stateCluster := range stateClusters {
			_clusters[stateCluster.Name] = &cluster{
				address:       stateCluster.Address,
				googleProject: stateCluster.GoogleProject,
				location:      *stateCluster.Location,
				releases:      make(map[string]*release),
				destination: destination{
					name:             stateCluster.Name,
					base:             stateCluster.Base,
					requiredRole:     stateCluster.RequiredRole,
					destinationType:  terra.ClusterDestination,
					terraHelmfileRef: *stateCluster.HelmfileRef,
				},
			}
		}

		_environments := make(map[string]*environment)
		for _, stateEnvironment := range stateEnvironments {
			if _, knownCluster := _clusters[stateEnvironment.DefaultCluster]; stateEnvironment.DefaultCluster != "" && !knownCluster {
				log.Warn().Msgf("environment '%s' had cluster '%s' that we do not have: race condition detected, retrying...",
					stateEnvironment.Name, stateEnvironment.DefaultCluster)
				continue retry
			}
			var lifecycle terra.Lifecycle
			if err := lifecycle.FromString(*stateEnvironment.Lifecycle); err != nil {
				return nil, err
			}
			var envAutoDelete autoDelete
			if !stateEnvironment.DeleteAfter.IsZero() {
				envAutoDelete.enabled = true
				envAutoDelete.after = time.Time(stateEnvironment.DeleteAfter)
			}
			var offline bool
			if stateEnvironment.Offline != nil {
				offline = *stateEnvironment.Offline
			}

			_environments[stateEnvironment.Name] = &environment{
				createdAt:                   time.Time(stateEnvironment.CreatedAt),
				defaultCluster:              _clusters[stateEnvironment.DefaultCluster],
				defaultNamespace:            stateEnvironment.DefaultNamespace,
				releases:                    make(map[string]*release),
				lifecycle:                   lifecycle,
				template:                    stateEnvironment.TemplateEnvironment,
				baseDomain:                  *stateEnvironment.BaseDomain,
				namePrefixesDomain:          *stateEnvironment.NamePrefixesDomain,
				uniqueResourcePrefix:        stateEnvironment.UniqueResourcePrefix,
				owner:                       stateEnvironment.Owner,
				preventDeletion:             *stateEnvironment.PreventDeletion,
				autoDelete:                  envAutoDelete,
				offline:                     offline,
				offlineScheduleBeginEnabled: stateEnvironment.OfflineScheduleBeginEnabled,
				offlineScheduleBeginTime:    time.Time(stateEnvironment.OfflineScheduleBeginTime),
				offlineScheduleEndEnabled:   stateEnvironment.OfflineScheduleEndEnabled,
				offlineScheduleEndTime:      time.Time(stateEnvironment.OfflineScheduleEndTime),
				offlineScheduleEndWeekends:  stateEnvironment.OfflineScheduleEndWeekends,
				enableJanitor:               stateEnvironment.EnableJanitor,
				destination: destination{
					name:             stateEnvironment.Name,
					base:             stateEnvironment.Base,
					requiredRole:     stateEnvironment.RequiredRole,
					destinationType:  terra.EnvironmentDestination,
					terraHelmfileRef: *stateEnvironment.HelmfileRef,
				},
			}
		}

		for _, stateRelease := range stateReleases {
			if _, knownCluster := _clusters[stateRelease.Cluster]; stateRelease.Cluster != "" && !knownCluster {
				log.Warn().Msgf("chart release '%s' has cluster '%s' that we do not have: race condition detected, retrying...",
					stateRelease.Name, stateRelease.Cluster)
				continue retry
			}
			if _, knownEnvironment := _environments[stateRelease.Environment]; stateRelease.Environment != "" && !knownEnvironment {
				log.Warn().Msgf("chart release '%s' has environment '%s' that we do not have: race condition detected, retrying...",
					stateRelease.Name, stateRelease.Environment)
				continue retry
			}
			switch stateRelease.DestinationType {
			case "cluster":
				_clusters[stateRelease.Cluster].releases[stateRelease.Name] = &release{
					name:         stateRelease.Name,
					enabled:      true,
					releaseType:  terra.ClusterReleaseType,
					chartVersion: stateRelease.ChartVersionExact,
					chartName:    stateRelease.Chart,
					repo:         *stateRelease.ChartInfo.ChartRepo,
					namespace:    stateRelease.Namespace,
					cluster:      _clusters[stateRelease.Cluster],
					destination:  _clusters[stateRelease.Cluster],
					helmfileRef:  *stateRelease.HelmfileRef,
					appVersion:   stateRelease.AppVersionExact,
				}
			case "environment":
				var helmfileOverlays []string
				if e, present := _environments[stateRelease.Environment]; present && e.offline {
					helmfileOverlays = []string{"offline"}
				}
				_environments[stateRelease.Environment].releases[stateRelease.Name] = &release{
					name:             stateRelease.Name,
					enabled:          true,
					releaseType:      terra.AppReleaseType,
					chartVersion:     stateRelease.ChartVersionExact,
					chartName:        stateRelease.Chart,
					repo:             *stateRelease.ChartInfo.ChartRepo,
					namespace:        stateRelease.Namespace,
					cluster:          _clusters[stateRelease.Cluster],
					destination:      _environments[stateRelease.Environment],
					helmfileRef:      *stateRelease.HelmfileRef,
					helmfileOverlays: helmfileOverlays,
					appVersion:       stateRelease.AppVersionExact,
					subdomain:        stateRelease.Subdomain,
					protocol:         stateRelease.Protocol,
					port:             int(stateRelease.Port),
				}
			default:
				return nil, errors.Errorf("unexpected destination type '%s' for release '%s'", stateRelease.DestinationType, stateRelease.Name)
			}
		}

		_state := &state{
			sherlock:     s.sherlock,
			environments: _environments,
			clusters:     _clusters,
		}
		s.cached = _state
		return _state, nil
	}
	return nil, errors.Errorf("ran out of retries trying to resolve race conditions while loading state from sherlock")
}

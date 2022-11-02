package sherlock

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
)

type stateLoader struct {
	sherlock    sherlock.StateReadWriter
	shellRunner shell.Runner
	thelmaHome  string
}

func NewStateLoader(thelmaHome string, shellRunner shell.Runner, sherlock sherlock.StateReadWriter) terra.StateLoader {
	return &stateLoader{
		thelmaHome:  thelmaHome,
		shellRunner: shellRunner,
		sherlock:    sherlock,
	}
}

func (s *stateLoader) Load() (terra.State, error) {
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

		clusters := make(map[string]*cluster)
		for _, stateCluster := range stateClusters {
			clusters[stateCluster.Name] = &cluster{
				address:       stateCluster.Address,
				googleProject: stateCluster.GoogleProject,
				location:      *stateCluster.Location,
				releases:      make(map[string]*clusterRelease),
				destination: destination{
					name:             stateCluster.Name,
					base:             stateCluster.Base,
					requireSuitable:  *stateCluster.RequiresSuitability,
					destinationType:  terra.ClusterDestination,
					terraHelmfileRef: *stateCluster.HelmfileRef,
				},
			}
		}

		environments := make(map[string]*environment)
		for _, stateEnvironment := range stateEnvironments {
			if _, knownCluster := clusters[stateEnvironment.DefaultCluster]; stateEnvironment.DefaultCluster != "" && !knownCluster {
				log.Warn().Msgf("environment '%s' had cluster '%s' that we do not have: race condition detected, retrying...",
					stateEnvironment.Name, stateEnvironment.DefaultCluster)
				continue retry
			}
			var lifecycle terra.Lifecycle
			if err := lifecycle.FromString(*stateEnvironment.Lifecycle); err != nil {
				return nil, err
			}
			environments[stateEnvironment.Name] = &environment{
				defaultCluster:       clusters[stateEnvironment.DefaultCluster],
				defaultNamespace:     stateEnvironment.DefaultNamespace,
				releases:             make(map[string]*appRelease),
				lifecycle:            lifecycle,
				template:             stateEnvironment.TemplateEnvironment,
				baseDomain:           *stateEnvironment.BaseDomain,
				namePrefixesDomain:   *stateEnvironment.NamePrefixesDomain,
				uniqueResourcePrefix: stateEnvironment.UniqueResourcePrefix,
				destination: destination{
					name:             stateEnvironment.Name,
					base:             stateEnvironment.Base,
					requireSuitable:  *stateEnvironment.RequiresSuitability,
					destinationType:  terra.EnvironmentDestination,
					terraHelmfileRef: *stateEnvironment.HelmfileRef,
				},
			}
		}

		for _, stateRelease := range stateReleases {
			if _, knownCluster := clusters[stateRelease.Cluster]; stateRelease.Cluster != "" && !knownCluster {
				log.Warn().Msgf("chart release '%s' has cluster '%s' that we do not have: race condition detected, retrying...",
					stateRelease.Name, stateRelease.Cluster)
				continue retry
			}
			if _, knownEnvironment := environments[stateRelease.Environment]; stateRelease.Environment != "" && !knownEnvironment {
				log.Warn().Msgf("chart release '%s' has environment '%s' that we do not have: race condition detected, retrying...",
					stateRelease.Name, stateRelease.Environment)
				continue retry
			}
			switch stateRelease.DestinationType {
			case "cluster":
				clusters[stateRelease.Cluster].releases[stateRelease.Name] = &clusterRelease{
					release: release{
						name:                stateRelease.Name,
						enabled:             true,
						releaseType:         terra.ClusterReleaseType,
						chartVersion:        stateRelease.ChartVersionExact,
						chartName:           stateRelease.Chart,
						repo:                *stateRelease.ChartInfo.ChartRepo,
						namespace:           stateRelease.Namespace,
						cluster:             clusters[stateRelease.Cluster],
						destination:         clusters[stateRelease.Cluster],
						helmfileRef:         *stateRelease.HelmfileRef,
						firecloudDevelopRef: stateRelease.FirecloudDevelopRef,
					},
				}
			case "environment":
				environments[stateRelease.Environment].releases[stateRelease.Name] = &appRelease{
					appVersion: stateRelease.AppVersionExact,
					subdomain:  stateRelease.Subdomain,
					protocol:   stateRelease.Protocol,
					port:       int(stateRelease.Port),
					release: release{
						name:                stateRelease.Name,
						enabled:             true,
						releaseType:         terra.AppReleaseType,
						chartVersion:        stateRelease.ChartVersionExact,
						chartName:           stateRelease.Chart,
						repo:                *stateRelease.ChartInfo.ChartRepo,
						namespace:           stateRelease.Namespace,
						cluster:             clusters[stateRelease.Cluster],
						destination:         environments[stateRelease.Environment],
						helmfileRef:         *stateRelease.HelmfileRef,
						firecloudDevelopRef: stateRelease.FirecloudDevelopRef,
					},
				}
			default:
				return nil, fmt.Errorf("unexpected destination type '%s' for release '%s'", stateRelease.DestinationType, stateRelease.Name)
			}
		}

		return &state{
			sherlock:     s.sherlock,
			environments: environments,
			clusters:     clusters,
		}, nil
	}
	return nil, fmt.Errorf("ran out of retries trying to resolve race conditions while loading state from sherlock")
}

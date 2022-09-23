package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

type stateLoader struct {
	sherlock    sherlock.StateLoader
	shellRunner shell.Runner
	thelmaHome  string
}

func NewStateLoader(thelmaHome string, shellRunner shell.Runner, sherlock sherlock.StateLoader) terra.StateLoader {
	return &stateLoader{
		thelmaHome:  thelmaHome,
		shellRunner: shellRunner,
		sherlock:    sherlock,
	}
}

func (s *stateLoader) Load() (terra.State, error) {
	stateClusters, err := s.sherlock.Clusters()
	if err != nil {
		return nil, err
	}

	stateEnvironments, err := s.sherlock.Environments()
	if err != nil {
		return nil, err
	}

	// transforms cluster data returned by sherlock client into thelma's state domain types
	clusters, err := s.buildClustersState(stateClusters)
	if err != nil {
		return nil, err
	}

	// transforms environment data returned by sherlock client into thelma's state domain types
	environments, err := s.buildEnvironmentsState(stateEnvironments, clusters)
	if err != nil {
		return nil, err
	}

	return &state{
		sherlock:     s.sherlock,
		environments: environments,
		clusters:     clusters,
	}, nil
}

func (s *stateLoader) buildClustersState(clusters sherlock.Clusters) (map[string]*cluster, error) {
	result := make(map[string]*cluster)
	for _, cl := range clusters {
		releases := make(map[string]*clusterRelease)
		c := &cluster{
			address:       cl.Address,
			googleProject: cl.GoogleProject,
			destination: destination{
				name:            cl.Name,
				base:            cl.Base,
				requireSuitable: *cl.RequiresSuitability,
				destinationType: terra.ClusterDestination,
			},
		}
		stateReleases, err := s.sherlock.ClusterReleases(cl.Name)
		if err != nil {
			return nil, err
		}

		for _, r := range stateReleases {
			releases[r.Name] = &clusterRelease{
				release: release{
					name:         r.Name,
					enabled:      true,
					releaseType:  terra.ClusterReleaseType,
					chartVersion: r.ChartVersionExact,
					chartName:    r.Chart,
					repo:         *r.ChartInfo.ChartRepo,
					namespace:    r.Namespace,
					helmfileRef:  r.HelmfileRefOrDefault("HEAD"),
					cluster:      c,
					destination:  c,
				},
			}
		}
		c.releases = releases
		result[c.Name()] = c
	}
	return result, nil
}

func (s *stateLoader) buildEnvironmentsState(environments sherlock.Environments, clusters map[string]*cluster) (map[string]*environment, error) {
	result := make(map[string]*environment)
	for _, env := range environments {
		releases := make(map[string]*appRelease)
		var lifecycle terra.Lifecycle
		if err := lifecycle.FromString(*env.Lifecycle); err != nil {
			return nil, err
		}
		e := &environment{
			defaultCluster:     clusters[env.DefaultCluster],
			lifecycle:          lifecycle,
			template:           env.TemplateEnvironment,
			baseDomain:         *env.BaseDomain,
			namePrefixesDomain: *env.NamePrefixesDomain,
			destination: destination{
				name:            env.Name,
				base:            env.Base,
				requireSuitable: *env.RequiresSuitability,
				destinationType: terra.EnvironmentDestination,
			},
		}
		stateReleases, err := s.sherlock.EnvironmentReleases(env.Name)
		if err != nil {
			return nil, err
		}

		for _, r := range stateReleases {
			releases[r.Name] = &appRelease{
				appVersion: r.AppVersionExact,
				subdomain:  r.Subdomain,
				protocol:   r.Protocol,
				port:       int(r.Port),
				release: release{
					name:         r.Name,
					enabled:      true,
					releaseType:  terra.AppReleaseType,
					chartVersion: r.ChartVersionExact,
					chartName:    r.Chart,
					repo:         *r.ChartInfo.ChartRepo,
					namespace:    r.Namespace,
					helmfileRef:  *r.HelmfileRef,
					cluster:      clusters[env.DefaultCluster],
					destination:  e,
				},
			}
		}
		e.releases = releases
		result[e.Name()] = e
	}

	return result, nil
}

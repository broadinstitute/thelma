package sherlock

import (
	"fmt"
	"strings"

	"github.com/broadinstitute/sherlock/clients/go/client/chart_releases"
	"github.com/broadinstitute/sherlock/clients/go/client/charts"
	"github.com/broadinstitute/sherlock/clients/go/client/clusters"
	"github.com/broadinstitute/sherlock/clients/go/client/environments"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
)

// WriteEnvironments will take a list of terra.Environment interfaces them and issue POST requests
// to write both the environment and any releases within that environment. 409 Conflict responses are ignored
func (s *Client) WriteEnvironments(envs []terra.Environment) ([]string, error) {
	createdEnvNames := make([]string, 0)
	for _, environment := range envs {
		log.Info().Msgf("exporting state for environment: %s", environment.Name())
		newEnv := toModelCreatableEnvironment(environment)

		newEnvRequestParams := environments.NewPostAPIV2EnvironmentsParams().
			WithEnvironment(newEnv)
		_, createdEnv, err := s.client.Environments.PostAPIV2Environments(newEnvRequestParams)
		var envAlreadyExists bool
		if err != nil {
			// Don't error if creating the chart results in 409 conflict
			if _, ok := err.(*environments.PostAPIV2EnvironmentsConflict); !ok {
				return nil, fmt.Errorf("error creating cluster: %v", err)
			}
			envAlreadyExists = true
		}

		// extract the generated name from a new dynamic environment
		var envName string
		if environment.Lifecycle().IsDynamic() && !envAlreadyExists {
			envName = createdEnv.Payload.Name
		} else {
			envName = environment.Name()
		}

		log.Debug().Msgf("environment name: %s", envName)
		if err := s.writeReleases(envName, environment.Releases()); err != nil {
			return nil, err
		}
		createdEnvNames = append(createdEnvNames, envName)
	}
	return createdEnvNames, nil
}

// WriteClusters will take a list of terra.Cluster interfaces them and issue POST requests
// to create both the cluster and any releases within that cluster. 409 Conflict responses are ignored
func (s *Client) WriteClusters(cls []terra.Cluster) error {
	for _, cluster := range cls {
		log.Info().Msgf("exporting state for cluster: %s", cluster.Name())
		newCluster := toModelCreatableCluster(cluster)
		newClusterRequestParams := clusters.NewPostAPIV2ClustersParams().
			WithCluster(newCluster)
		_, _, err := s.client.Clusters.PostAPIV2Clusters(newClusterRequestParams)
		if err != nil {
			// Don't error if creating the chart results in 409 conflict
			if _, ok := err.(*clusters.PostAPIV2ClustersConflict); !ok {
				return fmt.Errorf("error creating cluster: %v", err)
			}
		}

		if err := s.writeReleases(cluster.Name(), cluster.Releases()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Client) DeleteEnvironments(envs []terra.Environment) ([]string, error) {
	deletedEnvs := make([]string, 0)
	for _, env := range envs {
		// delete chart releases associated with environment
		releases := env.Releases()
		for _, release := range releases {
			if err := s.deleteRelease(release); err != nil {
				log.Warn().Msgf("error deleting chart release %s in environment %s: %v", release.Name(), env.Name(), err)
			}
		}
		params := environments.NewDeleteAPIV2EnvironmentsSelectorParams().
			WithSelector(env.Name())

		deletedEnv, err := s.client.Environments.DeleteAPIV2EnvironmentsSelector(params)
		if err != nil {
			return nil, fmt.Errorf("error deleting environment %s: %v", env.Name(), err)
		}
		log.Debug().Msgf("%#v", deletedEnv)
		deletedEnvs = append(deletedEnvs, deletedEnv.Payload.Name)
	}
	return deletedEnvs, nil
}

func toModelCreatableEnvironment(env terra.Environment) *models.V2controllersCreatableEnvironment {
	return &models.V2controllersCreatableEnvironment{
		Base:                env.Base(),
		BaseDomain:          utils.Nullable(env.BaseDomain()),
		DefaultCluster:      env.DefaultCluster().Name(),
		DefaultNamespace:    env.Namespace(),
		Lifecycle:           utils.Nullable(env.Lifecycle().String()),
		Name:                env.Name(),
		NamePrefixesDomain:  utils.Nullable(env.NamePrefixesDomain()),
		RequiresSuitability: utils.Nullable(env.RequireSuitable()),
		TemplateEnvironment: env.Template(),
	}
}

func toModelCreatableCluster(cluster terra.Cluster) *models.V2controllersCreatableCluster {
	// Hard coding to google for now since we don't have azure clusters
	provider := "google"
	return &models.V2controllersCreatableCluster{
		Address:             cluster.Address(),
		Base:                cluster.Base(),
		Name:                cluster.Name(),
		Provider:            &provider,
		GoogleProject:       cluster.Project(),
		RequiresSuitability: utils.Nullable(cluster.RequireSuitable()),
	}
}

func (s *Client) writeReleases(destinationName string, releases []terra.Release) error {
	// for each release attempt to create a chart
	for _, release := range releases {
		log.Info().Msgf("exporting release: %v", release.Name())
		// attempt to convert to app release
		if release.IsAppRelease() {
			appRelease := release.(terra.AppRelease)
			if err := s.writeAppRelease(destinationName, appRelease); err != nil {
				return err
			}
		} else if release.IsClusterRelease() {
			clusterRelease := release.(terra.ClusterRelease)
			if err := s.writeClusterRelease(clusterRelease); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Client) writeAppRelease(environmentName string, release terra.AppRelease) error {
	modelChart := models.V2controllersCreatableChart{
		Name:            release.ChartName(),
		ChartRepo:       utils.Nullable(release.Repo()),
		DefaultPort:     utils.Nullable(int64(release.Port())),
		DefaultProtocol: utils.Nullable(release.Protocol()),
	}
	// first try to create the chart
	newChartRequestParams := charts.NewPostAPIV2ChartsParams().
		WithChart(&modelChart)

	_, _, err := s.client.Charts.PostAPIV2Charts(newChartRequestParams)
	if err != nil {
		// Don't error if creating the chart results in 409 conflict
		if _, ok := err.(*charts.PostAPIV2ChartsConflict); !ok {
			return fmt.Errorf("error creating chart: %v", err)
		}
	}
	// then the chart release
	releaseName := strings.Join([]string{release.ChartName(), environmentName}, "-")
	var releaseNamespace string
	if release.Environment().Lifecycle().IsDynamic() {
		releaseNamespace = environmentName
	} else {
		releaseNamespace = release.Namespace()
	}
	modelChartRelease := models.V2controllersCreatableChartRelease{
		AppVersionExact:   release.AppVersion(),
		Chart:             release.ChartName(),
		ChartVersionExact: release.ChartVersion(),
		Environment:       environmentName,
		HelmfileRef:       utils.Nullable("master"),
		Name:              releaseName,
		Namespace:         releaseNamespace,
		Port:              int64(release.Port()),
		Protocol:          release.Protocol(),
		Subdomain:         release.Subdomain(),
	}

	newChartReleaseRequestParams := chart_releases.NewPostAPIV2ChartReleasesParams().
		WithChartRelease(&modelChartRelease)

	_, _, err = s.client.ChartReleases.PostAPIV2ChartReleases(newChartReleaseRequestParams)
	if err != nil {
		if _, ok := err.(*chart_releases.PostAPIV2ChartReleasesConflict); !ok {
			return fmt.Errorf("error creating chart release: %v", err)
		}
	}
	return nil
}

func (s *Client) writeClusterRelease(release terra.ClusterRelease) error {
	modelChart := models.V2controllersCreatableChart{
		Name:            release.ChartName(),
		ChartRepo:       utils.Nullable(release.Repo()),
		DefaultPort:     nil,
		DefaultProtocol: nil,
	}

	// first try to create the chart
	newChartRequestParams := charts.NewPostAPIV2ChartsParams().
		WithChart(&modelChart)

	_, _, err := s.client.Charts.PostAPIV2Charts(newChartRequestParams)
	if err != nil {
		// Don't error if creating the chart results in 409 conflict
		if _, ok := err.(*charts.PostAPIV2ChartsConflict); !ok {
			return fmt.Errorf("error creating chart: %v", err)
		}
	}

	// then the chart release
	releaseName := strings.Join([]string{release.ChartName(), release.Cluster().Name()}, "-")
	modelChartRelease := models.V2controllersCreatableChartRelease{
		Chart:             release.ChartName(),
		ChartVersionExact: release.ChartVersion(),
		Cluster:           release.ClusterName(),
		HelmfileRef:       utils.Nullable("master"),
		Name:              releaseName,
		Namespace:         release.Namespace(),
	}

	newChartReleaseRequestParams := chart_releases.NewPostAPIV2ChartReleasesParams().
		WithChartRelease(&modelChartRelease)

	_, _, err = s.client.ChartReleases.PostAPIV2ChartReleases(newChartReleaseRequestParams)
	if err != nil {
		if _, ok := err.(*chart_releases.PostAPIV2ChartReleasesConflict); !ok {
			return fmt.Errorf("error creating chart release: %v", err)
		}
	}
	return nil
}

func (s *Client) deleteRelease(release terra.Release) error {
	params := chart_releases.NewDeleteAPIV2ChartReleasesSelectorParams().
		WithSelector(strings.Join([]string{release.ChartName(), release.Destination().Name()}, "-"))
	_, err := s.client.ChartReleases.DeleteAPIV2ChartReleasesSelector(params)
	return err
}

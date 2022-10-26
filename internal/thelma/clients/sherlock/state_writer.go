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

func (s *Client) EnableRelease(env terra.Environment, releaseName string) error {
	// need to pull info about the template env in order to set chart and app versions
	templateEnv, err := s.getEnvironment(env.Template())
	if err != nil {
		return fmt.Errorf("unable to fetch template %s: %v", env.Template(), err)
	}
	templateEnvName := templateEnv.Name
	// now look up the chart release to enable in the template
	templateRelease, err := s.getChartRelease(templateEnvName, releaseName)
	if err != nil {
		return fmt.Errorf("unable to enable release, error retrieving from template: %v", err)
	}

	// enable the chart-release in environment
	enabledChart := &models.V2controllersCreatableChartRelease{
		AppVersionExact:     templateRelease.AppVersionExact,
		Chart:               templateRelease.Chart,
		ChartVersionExact:   templateRelease.ChartVersionExact,
		Environment:         env.Name(),
		HelmfileRef:         templateRelease.HelmfileRef,
		Port:                templateRelease.Port,
		Protocol:            templateRelease.Protocol,
		Subdomain:           templateRelease.Subdomain,
		FirecloudDevelopRef: templateRelease.FirecloudDevelopRef,
	}
	log.Info().Msgf("enabling chart-release: %q in environment: %q", releaseName, env.Name())
	params := chart_releases.NewPostAPIV2ChartReleasesParams().WithChartRelease(enabledChart)
	_, _, err = s.client.ChartReleases.PostAPIV2ChartReleases(params)
	return err
}

func (s *Client) DisableRelease(envName, releaseName string) error {
	params := chart_releases.NewDeleteAPIV2ChartReleasesSelectorParams().WithSelector(strings.Join([]string{envName, releaseName}, "/"))
	_, err := s.client.ChartReleases.DeleteAPIV2ChartReleasesSelector(params)
	return err
}

func toModelCreatableEnvironment(env terra.Environment) *models.V2controllersCreatableEnvironment {
	// if Helmfile ref isn't set it should default to head
	var helmfileRef string
	if env.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = env.TerraHelmfileRef()
	}
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
		HelmfileRef:         utils.Nullable(helmfileRef),
	}
}

func toModelCreatableCluster(cluster terra.Cluster) *models.V2controllersCreatableCluster {
	// Hard coding to google for now since we don't have azure clusters
	provider := "google"
	// if Helmfile ref isn't set it should default to head
	var helmfileRef string
	if cluster.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = cluster.TerraHelmfileRef()
	}
	return &models.V2controllersCreatableCluster{
		Address:             cluster.Address(),
		Base:                cluster.Base(),
		Name:                cluster.Name(),
		Provider:            &provider,
		GoogleProject:       cluster.Project(),
		RequiresSuitability: utils.Nullable(cluster.RequireSuitable()),
		HelmfileRef:         &helmfileRef,
		Location:            utils.Nullable(cluster.Location()),
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
	log.Debug().Msgf("release name: %v", release.Name())
	modelChart := models.V2controllersCreatableChart{
		Name:            release.ChartName(),
		ChartRepo:       utils.Nullable(release.Repo()),
		DefaultPort:     utils.Nullable(int64(release.Port())),
		DefaultProtocol: utils.Nullable(release.Protocol()),
		// TODO don't default this figure out how thelma actually determines if legacy configs should be rendered
		LegacyConfigsEnabled: utils.Nullable(true),
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
	// check for a release name override
	var releaseName string
	if release.Name() == release.ChartName() {
		releaseName = strings.Join([]string{release.ChartName(), environmentName}, "-")
	} else {
		releaseName = strings.Join([]string{release.Name(), environmentName}, "-")
	}

	// helmfile ref should default to HEAD if unspecified
	var helmfileRef string
	if release.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = release.TerraHelmfileRef()
	}

	modelChartRelease := models.V2controllersCreatableChartRelease{
		AppVersionExact:     release.AppVersion(),
		Chart:               release.ChartName(),
		ChartVersionExact:   release.ChartVersion(),
		Cluster:             release.ClusterName(),
		Environment:         environmentName,
		HelmfileRef:         utils.Nullable(helmfileRef),
		Name:                releaseName,
		Namespace:           release.Namespace(),
		Port:                int64(release.Port()),
		Protocol:            release.Protocol(),
		Subdomain:           release.Subdomain(),
		FirecloudDevelopRef: release.FirecloudDevelopRef(),
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
		// Cluster releases will never have legacy configs
		LegacyConfigsEnabled: utils.Nullable(false),
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

	// check for a release name override
	var releaseName string
	if release.Name() == release.ChartName() {
		releaseName = strings.Join([]string{release.ChartName(), release.ClusterName()}, "-")
	} else {
		releaseName = strings.Join([]string{release.Name(), release.ClusterName()}, "-")
	}
	// helmfile ref should default to HEAD if unspecified
	var helmfileRef string
	if release.TerraHelmfileRef() == "" {
		helmfileRef = "HEAD"
	} else {
		helmfileRef = release.TerraHelmfileRef()
	}
	modelChartRelease := models.V2controllersCreatableChartRelease{
		Chart:               release.ChartName(),
		ChartVersionExact:   release.ChartVersion(),
		Cluster:             release.ClusterName(),
		HelmfileRef:         utils.Nullable(helmfileRef),
		Name:                releaseName,
		Namespace:           release.Namespace(),
		FirecloudDevelopRef: release.FirecloudDevelopRef(),
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

func (s *Client) getEnvironment(name string) (*Environment, error) {
	params := environments.NewGetAPIV2EnvironmentsSelectorParams().WithSelector(name)
	environment, err := s.client.Environments.GetAPIV2EnvironmentsSelector(params)
	if err != nil {
		return nil, err
	}

	return &Environment{environment.Payload}, nil
}

func (s *Client) getChartRelease(environmentName, releaseName string) (*Release, error) {
	params := chart_releases.NewGetAPIV2ChartReleasesSelectorParams().WithSelector(strings.Join([]string{environmentName, releaseName}, "/"))
	release, err := s.client.ChartReleases.GetAPIV2ChartReleasesSelector(params)
	if err != nil {
		return nil, err
	}
	return &Release{release.Payload}, nil
}

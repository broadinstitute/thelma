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
)

type StateWriter interface {
	WriteEnvironments([]terra.Environment) error
	WriteClusters([]terra.Cluster) error
}

func (s *Client) WriteEnvironments(envs []terra.Environment) error {
	for _, environment := range envs {
		newEnv := toModelCreatableEnvironment(environment)

		newEnvRequestParams := environments.NewPostAPIV2EnvironmentsParams().
			WithEnvironment(newEnv)
		_, _, err := s.client.Environments.PostAPIV2Environments(newEnvRequestParams)
		if err != nil {
			return err
		}

		if err := s.writeReleases(environment.Releases()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Client) WriteClusters(cls []terra.Cluster) error {
	for _, cluster := range cls {
		newCluster := toModelCreatableCluster(cluster)

		newClusterRequestParams := clusters.NewPostAPIV2ClustersParams().
			WithCluster(newCluster)
		_, _, err := s.client.Clusters.PostAPIV2Clusters(newClusterRequestParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func toModelCreatableEnvironment(env terra.Environment) *models.V2controllersCreatableEnvironment {
	// set function *T optional values since function returns are not directly
	// addressable via pointers
	baseDomain := env.BaseDomain()
	lifecycle := env.Lifecycle().String()
	namePrefixesDomain := env.NamePrefixesDomain()
	requireSuitability := env.RequireSuitable()

	return &models.V2controllersCreatableEnvironment{
		Base:       env.Base(),
		BaseDomain: &baseDomain,
		// ChartReleasesFromTemplate: environment,
		DefaultCluster:      env.DefaultCluster().Name(),
		DefaultNamespace:    env.Namespace(),
		Lifecycle:           &lifecycle,
		Name:                env.Name(),
		NamePrefixesDomain:  &namePrefixesDomain,
		RequiresSuitability: &requireSuitability,
		TemplateEnvironment: env.Template(),
	}
}

func toModelCreatableCluster(cluster terra.Cluster) *models.V2controllersCreatableCluster {
	// Hard coding to google for now since we don't have azure clusters
	provider := "google"
	requireSuitability := cluster.RequireSuitable()
	return &models.V2controllersCreatableCluster{
		Address:             cluster.Address(),
		Base:                cluster.Base(),
		Name:                cluster.Name(),
		Provider:            &provider,
		GoogleProject:       cluster.Project(),
		RequiresSuitability: &requireSuitability,
	}
}

func (s *Client) writeReleases(releases []terra.Release) error {
	// for each release attempt to create a chart
	for _, release := range releases {
		// attempt to convert to app release
		var r terra.AppRelease
		if release.IsAppRelease() {
			r = release.(terra.AppRelease)
		}

		chartRepo := release.Repo()
		port := int64(r.Port())
		protocol := r.Protocol()
		modelChart := models.V2controllersCreatableChart{
			Name:            r.ChartName(),
			ChartRepo:       &chartRepo,
			DefaultPort:     &port,
			DefaultProtocol: &protocol,
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
		releaseName := strings.Join([]string{r.ChartName(), r.Environment().Name()}, "-")
		modelChartRelease := models.V2controllersCreatableChartRelease{
			AppVersionExact:   r.AppVersion(),
			Chart:             r.ChartName(),
			ChartVersionExact: r.ChartVersion(),
			Environment:       r.Environment().Name(),
			HelmfileRef:       utils.Nullable("master"),
			Name:              releaseName,
			Namespace:         r.Namespace(),
			Port:              int64(r.Port()),
			Protocol:          r.Protocol(),
			Subdomain:         r.Subdomain(),
		}

		newChartReleaseRequestParams := chart_releases.NewPostAPIV2ChartReleasesParams().
			WithChartRelease(&modelChartRelease)
		_, _, err = s.client.ChartReleases.PostAPIV2ChartReleases(newChartReleaseRequestParams)
		if err != nil {
			return err
		}
	}

	return nil
}

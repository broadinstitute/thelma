package sherlock

import (
	"github.com/broadinstitute/sherlock/clients/go/client/clusters"
	"github.com/broadinstitute/sherlock/clients/go/client/environments"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
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

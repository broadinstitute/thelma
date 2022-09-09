package sherlock

import (
	"github.com/broadinstitute/sherlock/clients/go/client/chart_releases"
	"github.com/broadinstitute/sherlock/clients/go/client/clusters"
	"github.com/broadinstitute/sherlock/clients/go/client/environments"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
)

type StateLoader interface {
	Environments() ([]*models.V2controllersEnvironment, error)
	Clusters() ([]*models.V2controllersCluster, error)
	ClusterReleases(clusterName string) ([]*models.V2controllersChartRelease, error)
	EnvironmentReleases(environmentName string) ([]*models.V2controllersChartRelease, error)
}

func (c *Client) Environments() ([]*models.V2controllersEnvironment, error) {
	params := environments.NewGetAPIV2EnvironmentsParams()
	environmentsResponse, err := c.client.Environments.GetAPIV2Environments(params)
	if err != nil {
		return nil, err
	}

	environments := environmentsResponse.GetPayload()

	return environments, nil
}

func (c *Client) Clusters() ([]*models.V2controllersCluster, error) {
	params := clusters.NewGetAPIV2ClustersParams()
	clustersResponse, err := c.client.Clusters.GetAPIV2Clusters(params)
	if err != nil {
		return nil, err
	}

	clusters := clustersResponse.GetPayload()

	return clusters, nil
}

func (c *Client) ClusterReleases(clusterName string) ([]*models.V2controllersChartRelease, error) {
	params := chart_releases.NewGetAPIV2ChartReleasesParams()
	params.Cluster = &clusterName

	clusterReleasesResponse, err := c.client.ChartReleases.GetAPIV2ChartReleases(params)
	if err != nil {
		return nil, err
	}

	clusterReleases := clusterReleasesResponse.GetPayload()

	return clusterReleases, nil
}

func (c *Client) EnvironmentReleases(environmentName string) ([]*models.V2controllersChartRelease, error) {
	params := chart_releases.NewGetAPIV2ChartReleasesParams()
	params.Environment = &environmentName

	environmentReleasesResponse, err := c.client.ChartReleases.GetAPIV2ChartReleases(params)
	if err != nil {
		return nil, err
	}

	environmentReleases := environmentReleasesResponse.GetPayload()

	return environmentReleases, nil
}

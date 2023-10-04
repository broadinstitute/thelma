package sherlock

import (
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/chart_releases"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/clusters"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/environments"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
)

type StateLoader interface {
	Environments() (Environments, error)
	Clusters() (Clusters, error)
	Releases() (Releases, error)
}

type Cluster struct {
	*models.SherlockClusterV3
}

type Clusters []Cluster

type Environment struct {
	*models.V2controllersEnvironment
}

type Environments []Environment

type Release struct {
	*models.V2controllersChartRelease
}

type Releases []Release

func wrapEnvironments(envs ...*models.V2controllersEnvironment) Environments {
	environments := make([]Environment, 0)
	for _, env := range envs {
		environments = append(environments, Environment{env})
	}

	return environments
}

func (c *Client) Environments() (Environments, error) {
	params := environments.NewGetAPIV2EnvironmentsParams()
	environmentsResponse, err := c.client.Environments.GetAPIV2Environments(params)
	if err != nil {
		return nil, err
	}

	environmentsPayload := environmentsResponse.GetPayload()
	environments := wrapEnvironments(environmentsPayload...)

	return environments, nil
}

func wrapClusters(cls ...*models.SherlockClusterV3) Clusters {
	clusters := make([]Cluster, 0)
	for _, cluster := range cls {
		clusters = append(clusters, Cluster{cluster})
	}

	return clusters
}

func (c *Client) Clusters() (Clusters, error) {
	params := clusters.NewGetAPIClustersV3Params()
	clustersResponse, err := c.client.Clusters.GetAPIClustersV3(params)
	if err != nil {
		return nil, err
	}

	clustersPayload := clustersResponse.GetPayload()
	clusters := wrapClusters(clustersPayload...)

	return clusters, nil
}

func wrapReleases(rs ...*models.V2controllersChartRelease) Releases {
	releases := make([]Release, 0)
	for _, release := range rs {
		releases = append(releases, Release{release})
	}

	return releases
}

func (c *Client) Releases() (Releases, error) {
	response, err := c.client.ChartReleases.GetAPIV2ChartReleases(
		chart_releases.NewGetAPIV2ChartReleasesParams())
	if err != nil {
		return nil, err
	}
	return wrapReleases(response.Payload...), nil
}

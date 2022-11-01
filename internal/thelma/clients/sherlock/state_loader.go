package sherlock

import (
	"github.com/broadinstitute/sherlock/clients/go/client/chart_releases"
	"github.com/broadinstitute/sherlock/clients/go/client/clusters"
	"github.com/broadinstitute/sherlock/clients/go/client/environments"
	"github.com/broadinstitute/sherlock/clients/go/client/models"
)

type StateLoader interface {
	Environments() (Environments, error)
	Clusters() (Clusters, error)
	ClusterReleases(clusterName string) (Releases, error)
	EnvironmentReleases(environmentName string) (Releases, error)
	Releases() (Releases, error)
}

type Cluster struct {
	*models.V2controllersCluster
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

func wrapClusters(cls ...*models.V2controllersCluster) Clusters {
	clusters := make([]Cluster, 0)
	for _, cluster := range cls {
		clusters = append(clusters, Cluster{cluster})
	}

	return clusters
}

func (c *Client) Clusters() (Clusters, error) {
	params := clusters.NewGetAPIV2ClustersParams()
	clustersResponse, err := c.client.Clusters.GetAPIV2Clusters(params)
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

func (c *Client) ClusterReleases(clusterName string) (Releases, error) {
	desiredDestinationType := "cluster"
	params := chart_releases.NewGetAPIV2ChartReleasesParams()
	params.Cluster = &clusterName
	params.DestinationType = &desiredDestinationType

	clusterReleasesResponse, err := c.client.ChartReleases.GetAPIV2ChartReleases(params)
	if err != nil {
		return nil, err
	}

	clusterReleasesPayload := clusterReleasesResponse.GetPayload()
	clusterReleases := wrapReleases(clusterReleasesPayload...)

	return clusterReleases, nil
}

func (c *Client) EnvironmentReleases(environmentName string) (Releases, error) {
	params := chart_releases.NewGetAPIV2ChartReleasesParams()
	params.Environment = &environmentName

	environmentReleasesResponse, err := c.client.ChartReleases.GetAPIV2ChartReleases(params)
	if err != nil {
		return nil, err
	}

	environmentReleasesPayload := environmentReleasesResponse.GetPayload()
	environmentReleases := wrapReleases(environmentReleasesPayload...)

	return environmentReleases, nil
}

func (c *Client) Releases() (Releases, error) {
	response, err := c.client.ChartReleases.GetAPIV2ChartReleases(
		chart_releases.NewGetAPIV2ChartReleasesParams())
	if err != nil {
		return nil, err
	}
	return wrapReleases(response.Payload...), nil
}

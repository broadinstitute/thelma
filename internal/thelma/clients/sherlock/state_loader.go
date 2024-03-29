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
	*models.SherlockEnvironmentV3
}

type Environments []Environment

type Release struct {
	*models.SherlockChartReleaseV3
}

type Releases []Release

func wrapEnvironments(envs ...*models.SherlockEnvironmentV3) Environments {
	environments := make([]Environment, 0)
	for _, env := range envs {
		environments = append(environments, Environment{env})
	}

	return environments
}

func (c *clientImpl) Environments() (Environments, error) {
	environmentsResponse, err := c.client.Environments.GetAPIEnvironmentsV3(
		environments.NewGetAPIEnvironmentsV3Params())
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

func (c *clientImpl) Clusters() (Clusters, error) {
	params := clusters.NewGetAPIClustersV3Params()
	clustersResponse, err := c.client.Clusters.GetAPIClustersV3(params)
	if err != nil {
		return nil, err
	}

	clustersPayload := clustersResponse.GetPayload()
	clusters := wrapClusters(clustersPayload...)

	return clusters, nil
}

func wrapReleases(rs ...*models.SherlockChartReleaseV3) Releases {
	releases := make([]Release, 0)
	for _, release := range rs {
		releases = append(releases, Release{release})
	}

	return releases
}

func (c *clientImpl) Releases() (Releases, error) {
	response, err := c.client.ChartReleases.GetAPIChartReleasesV3(
		chart_releases.NewGetAPIChartReleasesV3Params())
	if err != nil {
		return nil, err
	}
	return wrapReleases(response.Payload...), nil
}

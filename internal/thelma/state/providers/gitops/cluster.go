package gitops

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

// Cluster represents a Terra cluster
type cluster struct {
	address  string // Cluster API address. Eg "https://10.0.0.1/api"
	releases map[string]*clusterRelease
	destination
}

// NewCluster constructs a new Cluster
func NewCluster(name string, base string, address string, releases map[string]*clusterRelease) terra.Cluster {
	return &cluster{
		address:  address,
		releases: releases,
		destination: destination{
			name:            name,
			base:            base,
			destinationType: terra.ClusterDestination,
		},
	}
}

func (c *cluster) Releases() []terra.Release {
	var result []terra.Release
	for _, r := range c.releases {
		if r.enabled {
			result = append(result, r)
		}
	}
	return result
}

func (c *cluster) Address() string {
	return c.address
}

func (c *cluster) ReleaseType() terra.ReleaseType {
	return terra.ClusterReleaseType
}

// Name cluster name, eg. "terra-alpha"
func (c *cluster) Name() string {
	return c.name
}

// Base cluster base, eg. "terra"
func (c *cluster) Base() string {
	return c.base
}

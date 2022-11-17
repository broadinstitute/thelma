package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"strings"
)

// Cluster represents a Terra cluster
type cluster struct {
	address  string // Cluster API address. Eg "https://10.0.0.1/api"
	project  string // Cluster Google project. Eg "broad-dsde-dev"
	location string // Cluster location. Eg. "us-central1-a"
	releases map[string]*clusterRelease
	destination
}

// newCluster constructs a new Cluster
func newCluster(name string, base string, address string, project string, location string, requireSuitable bool, releases map[string]*clusterRelease) *cluster {
	return &cluster{
		address:  address,
		project:  project,
		location: location,
		releases: releases,
		destination: destination{
			name:            name,
			base:            base,
			destinationType: terra.ClusterDestination,
			requireSuitable: requireSuitable,
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

func (c *cluster) Project() string {
	return c.project
}

func (c *cluster) ProjectSuffix() string {
	tokens := strings.Split(c.Project(), "-")
	return tokens[len(tokens)-1]
}

func (c *cluster) Location() string {
	return c.location
}

func (c *cluster) ReleaseType() terra.ReleaseType {
	return terra.ClusterReleaseType
}

func (c *cluster) ArtifactBucket() string {
	return fmt.Sprintf("thelma-artifacts-%s", c.name)
}

// Name cluster name, eg. "terra-alpha"
func (c *cluster) Name() string {
	return c.name
}

// Base cluster base, eg. "terra"
func (c *cluster) Base() string {
	return c.base
}

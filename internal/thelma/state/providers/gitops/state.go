// Package gitops is an implementation of the terra.State interface that reads configuration from terra-helmfile
package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statebucket"
)

// implements the terra.State interface
type state struct {
	statebucket  statebucket.StateBucket
	versions     Versions
	environments map[string]terra.Environment
	clusters     map[string]terra.Cluster
}

func (s *state) Destinations() terra.Destinations {
	return newDestinations(s)
}

func (s *state) Environments() terra.Environments {
	return newEnvironments(s)
}

func (s *state) Clusters() terra.Clusters {
	return newClusters(s)
}

func (s *state) Releases() terra.Releases {
	return newReleases(s)
}

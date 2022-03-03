package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/gitops/statebucket"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

type state struct {
	statebucket  statebucket.StateBucket
	versions     Versions
	environments map[string]terra.Environment
	clusters     map[string]terra.Cluster
}

func (g *state) Destinations() terra.Destinations {
	return newDestinations(g)
}

func (g *state) Environments() terra.Environments {
	return newEnvironments(g)
}

func (g *state) Clusters() terra.Clusters {
	return newClusters(g)
}

func (g *state) Releases() terra.Releases {
	return newReleases(g)
}

func Load(thelmaHome string, shellRunner shell.Runner) (terra.State, error) {
	_statebucket, err := statebucket.New()
	if err != nil {
		return nil, err
	}

	_versions, err := NewVersions(thelmaHome, shellRunner)
	if err != nil {
		return nil, err
	}

	_clusters, err := loadClusters(thelmaHome, _versions)
	if err != nil {
		return nil, err
	}

	_environments, err := loadEnvironments(thelmaHome, _versions, _clusters, _statebucket)
	if err != nil {
		return nil, err
	}

	return &state{
		statebucket:  _statebucket,
		versions:     _versions,
		clusters:     _clusters,
		environments: _environments,
	}, nil
}

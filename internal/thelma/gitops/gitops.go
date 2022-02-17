package gitops

import (
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

type gitops struct {
	versions     Versions
	environments map[string]terra.Environment
	clusters     map[string]terra.Cluster
}

func (g *gitops) Destinations() terra.Destinations {
	return newDestinations(g)
}

func (g *gitops) Environments() terra.Environments {
	return newEnvironments(g)
}

func (g *gitops) Releases() terra.Releases {
	return newReleases(g)
}

func Load(thelmaHome string, shellRunner shell.Runner) (terra.State, error) {
	_versions, err := NewVersions(thelmaHome, shellRunner)
	if err != nil {
		return nil, err
	}

	_clusters, err := loadClusters(thelmaHome, _versions)
	if err != nil {
		return nil, err
	}

	_environments, err := loadEnvironments(thelmaHome, _versions, _clusters)
	if err != nil {
		return nil, err
	}

	return &gitops{
		versions:     _versions,
		clusters:     _clusters,
		environments: _environments,
	}, nil
}

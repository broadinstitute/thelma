package stateval

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"strings"
)

// Cluster -- information about the cluster the release is being deployed to
type Cluster struct {
	// Name of the cluster this release is being deployed to (omitted if app release)
	Name string `yaml:"Name"`
	// GoogleProject name of the Google project the cluster lives in
	GoogleProject string `yaml:"GoogleProject"`
	// GoogleProjectSuffix suffix of the Google project where this release is being deployed
	GoogleProjectSuffix string `yaml:"GoogleProjectSuffix"`
}

func forCluster(cluster terra.Cluster) Cluster {
	return Cluster{
		Name:                cluster.Name(),
		GoogleProject:       cluster.Project(),
		GoogleProjectSuffix: projectSuffix(cluster),
	}
}

func projectSuffix(cluster terra.Cluster) string {
	tokens := strings.Split(cluster.Project(), "-")
	return tokens[len(tokens)-1]
}

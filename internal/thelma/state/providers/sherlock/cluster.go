package sherlock

import (
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

const clusterDefaultLocation = "us-central1-a"

type cluster struct {
	address       string
	googleProject string
	location      string
	releases      map[string]*clusterRelease
	destination
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
	return c.googleProject
}

func (c *cluster) ProjectSuffix() string {
	tokens := strings.Split(c.Project(), "-")
	return tokens[len(tokens)-1]
}

func (c *cluster) Location() string {
	if c.location == "" {
		return clusterDefaultLocation
	}
	return c.location
}

func (c *cluster) ReleaseType() terra.ReleaseType {
	return terra.ClusterReleaseType
}

func (c *cluster) Name() string {
	return c.name
}

func (c *cluster) Base() string {
	return c.base
}

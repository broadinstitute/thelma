package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type state struct {
	environments map[string]*environment
	clusters     map[string]*cluster
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

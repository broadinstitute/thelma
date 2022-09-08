package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type state struct {
	sherlock *sherlock.Client
	// versions     Versions
	environments map[string]*environment
	clusters     map[string]*cluster
}

func (s *state) Destinations() terra.Destinations {
	return newDestinations(s)
}

func (s *state) Environments() terra.Environments {
	return newEnvironments(s)
}

func (s *state) Clusers() terra.Clusters {
	return newClusters(s)
}

// func (s *state) Release() terra.Releases {

// }

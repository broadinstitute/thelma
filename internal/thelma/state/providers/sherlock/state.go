package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// StateReadWriter is an interface representing the ability to both read and
// create/update thelma's internal state using a sherlock client
type StateReadWriter interface {
	StateLoader
	sherlock.StateWriter
}

type state struct {
	sherlock     StateReadWriter
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

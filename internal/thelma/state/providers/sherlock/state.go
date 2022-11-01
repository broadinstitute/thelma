package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// state is an implementer of terra.State, the overall provider interface for Thelma
type state struct {
	sherlock     sherlock.StateReadWriter
	environments map[string]*environment
	clusters     map[string]*cluster
}

func (s *state) Destinations() terra.Destinations {
	return newDestinationsView(s)
}

func (s *state) Environments() terra.Environments {
	return newEnvironmentsView(s)
}

func (s *state) Clusters() terra.Clusters {
	return newClustersView(s)
}

func (s *state) Releases() terra.Releases {
	return newReleasesView(s)
}

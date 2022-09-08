package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
)

type state struct {
	sherlock *sherlock.Client
	// versions     Versions
	environments map[string]*environment
	clusters     map[string]*cluster
}

// func (s *state) Destinations() terra.Destinations {
// }

// func (s *state) Environments() terra.Environments {

// }

// func (s *state) Clusers() terra.Clusters {

// }

// func (s *state) Release() terra.Releases {

// }

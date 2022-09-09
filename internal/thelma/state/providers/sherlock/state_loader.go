package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

type stateLoader struct {
	sherlock    sherlock.StateLoader
	shellRunner shell.Runner
	thelmaHome  string
}

func NewStateLoader(thelmaHome string, shellRunner shell.Runner, sherlock sherlock.StateLoader) terra.StateLoader {
	return &stateLoader{
		thelmaHome:  thelmaHome,
		shellRunner: shellRunner,
		sherlock:    sherlock,
	}
}

func (s *stateLoader) Load() (terra.State, error) {
	panic("TODO")
}

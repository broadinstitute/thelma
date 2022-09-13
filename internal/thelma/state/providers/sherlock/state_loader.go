package sherlock

import (
	"github.com/broadinstitute/sherlock/clients/go/client/models"
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
	clusters, err := s.sherlock.Clusters()
	if err != nil {
		return nil, err
	}

	environments, err := s.sherlock.Environments()
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		_, err := s.sherlock.ClusterReleases(cluster.Name)
		if err != nil {
			return nil, err
		}
	}

	for _, environment := range environments {
		_, err := s.sherlock.EnvironmentReleases(environment.Name)
		if err != nil {
			return nil, err
		}
	}

	panic("TODO")
}

func toStateEnvironment(environment []*models.V2controllersEnvironment) terra.Environment

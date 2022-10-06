package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
)

type stateWriter struct {
	stateWriter terra.StateWriter
	state       terra.State
}

func NewSherlockStateWriter(state terra.State, writer terra.StateWriter) *stateWriter {
	return &stateWriter{
		stateWriter: writer,
		state:       state,
	}
}

func (s *stateWriter) WriteEnvironments() error {
	// Need to create template envs first as other envs that reference them will 404 in sherlock otherwise
	templateEnvfilter := filter.Environments().HasLifecycle(terra.Template)
	templateEnvs, err := s.state.Environments().Filter(templateEnvfilter)
	if err != nil {
		return err
	}
	if _, err := s.stateWriter.WriteEnvironments(templateEnvs); err != nil {
		return err
	}

	allOtherEnvsFilter := filter.Environments().
		Or(
			filter.Environments().HasLifecycle(terra.Dynamic),
			filter.Environments().HasLifecycle(terra.Static),
		)

	allOtherEnvs, err := s.state.Environments().Filter(allOtherEnvsFilter)

	if err != nil {
		return err
	}
	_, err = s.stateWriter.WriteEnvironments(allOtherEnvs)
	return err
}

func (s *stateWriter) WriteClusters() error {
	allClusters, err := s.state.Clusters().All()
	if err != nil {
		return err
	}
	return s.stateWriter.WriteClusters(allClusters)
}

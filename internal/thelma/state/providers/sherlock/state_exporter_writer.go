package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
)

type StateExporterWriter struct {
	stateWriter terra.StateWriter
	state       terra.State
}

func NewSherlockStateWriter(state terra.State, writer terra.StateWriter) *StateExporterWriter {
	return &StateExporterWriter{
		stateWriter: writer,
		state:       state,
	}
}

func (s *StateExporterWriter) WriteEnvironments() error {
	// Need to create template envs first as other envs that reference them will 404 in sherlock otherwise
	allEnvs, err := s.state.Environments().All()
	if err != nil {
		return err
	}

	isTemplate := filter.Environments().IsTemplate()

	templateEnvs := isTemplate.Filter(allEnvs)
	if _, err = s.stateWriter.WriteEnvironments(templateEnvs); err != nil {
		return err
	}

	allOtherEnvsFilter := isTemplate.Negate()

	allOtherEnvs := allOtherEnvsFilter.Filter(allEnvs)

	if err != nil {
		return err
	}
	_, err = s.stateWriter.WriteEnvironments(allOtherEnvs)
	return err
}

func (s *StateExporterWriter) WriteClusters() error {
	allClusters, err := s.state.Clusters().All()
	if err != nil {
		return err
	}
	return s.stateWriter.WriteClusters(allClusters)
}

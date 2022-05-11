package statebucket

import "fmt"

type schemaVerifier struct {
	inner         writer
	schemaVersion int32
}

func newSchemaVerifier(schemaVersion int32, w writer) writer {
	return &schemaVerifier{
		inner:         w,
		schemaVersion: schemaVersion,
	}
}

func (s *schemaVerifier) read() (StateFile, error) {
	state, err := s.inner.read()
	if err != nil {
		return state, err
	}
	err = s.checkSchemaVersion(state)
	return state, err
}

func (s *schemaVerifier) write(file StateFile) error {
	// write() means "clobber existing state with my content" so no check needed
	file.SchemaVersion = s.schemaVersion
	return s.inner.write(file)
}

func (s *schemaVerifier) update(fn transformFn) error {
	return s.inner.update(func(input StateFile) (output StateFile, err error) {
		if err := s.checkSchemaVersion(input); err != nil {
			return StateFile{}, err
		}
		// make sure the update is set to our schema version
		input.SchemaVersion = s.schemaVersion
		return fn(input)
	})
}

func (s *schemaVerifier) checkSchemaVersion(state StateFile) error {
	if state.SchemaVersion > s.schemaVersion {
		return fmt.Errorf("statefile schema version %d is greater than %d, the latest "+
			"schema support by this version of Thelma. Upgrade Thelma to use features that interact with the "+
			"state file", state.SchemaVersion, s.schemaVersion)
	}
	return nil
}

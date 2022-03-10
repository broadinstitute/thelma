package statebucket

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// fileWriter is used for testing

func newFileWriter(dir string) writer {
	return &fileWriter{dir: dir}
}

type fileWriter struct {
	dir string
}

func (f fileWriter) read() (StateFile, error) {
	content, err := os.ReadFile(f.filepath())
	if err != nil {
		return StateFile{}, fmt.Errorf("error reading %s: %v", f.filepath(), err)
	}

	var state StateFile
	if err := json.Unmarshal(content, &state); err != nil {
		return StateFile{}, fmt.Errorf("error unmarshalling %s: %v", f.filepath(), err)
	}
	return state, nil
}

func (f fileWriter) write(state StateFile) error {
	content, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("error marshalling state: %v", err)
	}
	return os.WriteFile(f.filepath(), content, 0600)
}

func (f fileWriter) update(fn transformFn) error {
	oldState, err := f.read()
	if err != nil {
		return err
	}
	newState, err := fn(oldState)
	if err != nil {
		return err
	}
	return f.write(newState)
}

func (f fileWriter) filepath() string {
	return path.Join(f.dir, stateObject)
}

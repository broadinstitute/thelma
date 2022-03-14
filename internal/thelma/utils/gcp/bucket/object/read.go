package object

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
)

// Read reads the content of an object in a GCS bucket
type Read interface {
	Content() []byte
	Operation
}

func NewRead() Read {
	return &read{}
}

type read struct {
	content []byte
}

func (r *read) Kind() string {
	return "read"
}

func (r *read) Content() []byte {
	return r.content
}

func (r *read) Handler(object Object, logger zerolog.Logger) error {
	reader, err := object.Handle.NewReader(object.Ctx)
	if err != nil {
		return fmt.Errorf("error reading object: %v", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading object: %v", err)
	}
	if err = reader.Close(); err != nil {
		return fmt.Errorf("error closing object reader: %v", err)
	}

	logTransfer(logger, int64(len(content)))
	r.content = content
	return nil
}

package object

import (
	"cloud.google.com/go/storage"
	"github.com/rs/zerolog"
)

type Exists interface {
	Exists() bool
	Operation
}

func NewExists() Exists {
	return &exists{}
}

type exists struct {
	exists bool
}

func (e *exists) Kind() string {
	return "exists"
}

func (e *exists) Exists() bool {
	return e.exists
}

func (e *exists) Handler(object Object, _ zerolog.Logger) error {
	_, err := object.Handle.Attrs(object.Ctx)
	if err == nil {
		e.exists = true
		return nil
	}
	if err == storage.ErrObjectNotExist {
		e.exists = false
		return nil
	}
	return err
}

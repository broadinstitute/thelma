package object

import (
	"cloud.google.com/go/storage"
	"github.com/rs/zerolog"
)

// Attrs reads the attributes of an object in a GCS bucket
type Attrs interface {
	Attrs() *storage.ObjectAttrs
	Operation
}

func NewAttrs() Attrs {
	return &attrs{}
}

type attrs struct {
	attrs *storage.ObjectAttrs
}

func (a *attrs) Kind() string {
	return "attrs"
}

func (a *attrs) Handler(object Object, _ zerolog.Logger) error {
	result, err := object.Handle.Attrs(object.Ctx)
	if err != nil {
		return err
	}
	a.attrs = result
	return nil
}

func (a *attrs) Attrs() *storage.ObjectAttrs {
	return a.attrs
}

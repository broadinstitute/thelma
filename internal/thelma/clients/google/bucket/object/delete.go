package object

import "github.com/rs/zerolog"

type Delete interface {
	SyncOperation
}

func NewDelete() Delete {
	return &_delete{}
}

type _delete struct{}

func (d *_delete) Kind() string {
	return "delete"
}

func (d *_delete) Handler(object Object, _ zerolog.Logger) error {
	return object.Handle.Delete(object.Ctx)
}

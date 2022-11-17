package object

import (
	"github.com/rs/zerolog"
)

// Update updates attributes of an object in a GCS bucket
type Update interface {
	SyncOperation
}

func NewUpdate(attrs AttrSet) Update {
	return &update{
		attrs: attrs,
	}
}

type update struct {
	attrs AttrSet
}

func (u *update) Kind() string {
	return "update"
}

func (u *update) Handler(object Object, logger zerolog.Logger) error {
	u.attrs.writeToLogEvent(logger.Debug())
	_, err := object.Handle.Update(object.Ctx, u.attrs.asUpdateAttrs())
	return err
}

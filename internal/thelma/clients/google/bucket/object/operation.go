package object

import (
	"github.com/rs/zerolog"
)

// Operation is an interface for operations on GCS bucket objects
type Operation interface {
	// Handler performs the call, given object and logger references
	Handler(object Object, logger zerolog.Logger) error
	// Kind returns the a description of the kind of this operation (eg. "delete", "upload")
	Kind() string
}

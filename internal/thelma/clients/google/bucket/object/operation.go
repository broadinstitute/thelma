package object

import (
	"github.com/rs/zerolog"
)

// SyncOperation is an interface for synchronous operations on GCS bucket objects
type SyncOperation interface {
	// Handler performs the call, given object and logger references
	Handler(object Object, logger zerolog.Logger) error
	// Kind returns the a description of the kind of this operation (eg. "delete", "upload")
	Kind() string
}

package lock

import (
	"cloud.google.com/go/storage"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/object"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Unlock interface {
	object.SyncOperation
}

func NewUnlock(generation int64) Unlock {
	return &unlock{
		generation: generation,
	}
}

type unlock struct {
	generation int64
}

func (u *unlock) Kind() string {
	return "unlock"
}

func (u *unlock) Handler(object object.Object, logger zerolog.Logger) error {
	logger = logger.With().Int64("generation", u.generation).Logger()

	withCondition := object.Handle.If(storage.Conditions{GenerationMatch: u.generation})
	if err := withCondition.Delete(object.Ctx); err != nil {
		if isPreconditionFailed(err) {
			logger.Warn().Msgf("Attempted to release lock, but another process has already claimed it")
			return nil
		}
		return errors.Errorf("error deleting lock: %v", err)
	}

	logger.Debug().Msgf("Successfully released lock")
	return nil
}

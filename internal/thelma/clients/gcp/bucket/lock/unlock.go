package lock

import (
	"cloud.google.com/go/storage"
	"fmt"
	object2 "github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/object"
	"github.com/rs/zerolog"
)

type Unlock interface {
	object2.Operation
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

func (u *unlock) Handler(object object2.Object, logger zerolog.Logger) error {
	logger = logger.With().Int64("generation", u.generation).Logger()

	withCondition := object.Handle.If(storage.Conditions{GenerationMatch: u.generation})
	if err := withCondition.Delete(object.Ctx); err != nil {
		if isPreconditionFailed(err) {
			logger.Warn().Msgf("Attempted to release lock, but another process has already claimed it")
			return nil
		}
		return fmt.Errorf("error deleting lock: %v", err)
	}

	logger.Debug().Msgf("Successfully released lock")
	return nil
}

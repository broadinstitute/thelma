package lock

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/object"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/googleapi"
	"net/http"
	"time"
)

// Options configuration parameters for a Locker
type Options struct {
	// ExpiresAfter will delete existing locks older than this age (0 means locks will never expire)
	ExpiresAfter time.Duration
	// BackoffStartingInterval initial sleep time for backoff
	BackoffStartingInterval time.Duration
	// BackoffMultiplier how much to multiply backoff interval after each failed attempt to acquire a lock
	BackoffMultiplier float64
	// MaxWait how long to wait for a lock before timing out
	MaxWait time.Duration
}

type Lock interface {
	Generation() int64
	object.SyncOperation
}

func NewLock(options Options) Lock {
	return &lock{
		options: options,
	}
}

type lock struct {
	options    Options
	generation int64
}

func (l *lock) Kind() string {
	return "lock"
}

func (l *lock) Generation() int64 {
	return l.generation
}

func (l *lock) Handler(object object.Object, logger zerolog.Logger) error {
	if l.options.ExpiresAfter > 0 {
		if err := l.deleteExpiredLock(object, logger); err != nil {
			return err
		}
	}
	return l.waitForLock(object, logger)
}

func (l *lock) waitForLock(object object.Object, logger zerolog.Logger) error {
	// only create the object if it does not already exist
	withCondition := object.Handle.If(storage.Conditions{DoesNotExist: true})

	ctx, cancelFn := context.WithTimeout(object.Ctx, l.options.MaxWait)
	defer cancelFn()

	backoff := l.options.BackoffStartingInterval
	attempt := 1

	for {
		logger := logger.With().Int("attempt", attempt).Logger()
		logger.Debug().Msgf("Attempting to obtain lock")

		// attempt to write empty object to bucket
		writer := withCondition.NewWriter(ctx)
		_, writeErr := writer.Write([]byte(""))
		closeErr := writer.Close()

		if writeErr == nil && closeErr == nil {
			// Success!
			generation := writer.Attrs().Generation

			logger.Debug().
				Int64("generation", generation).
				Msgf("Successfully obtained lock")

			l.generation = generation
			return nil
		}

		// We failed to grab the lock. Either someone else has it or something went wrong. Either way, retry after backoff
		logger = logger.With().Dur("retry-interval", backoff).Logger()
		if writeErr != nil {
			logger.Warn().
				Err(writeErr).
				Msgf("Unexpected error attempting to write to lock object: %v", writeErr)
		}
		if closeErr != nil {
			if isPreconditionFailed(closeErr) {
				logger.Debug().
					Msgf("Another process has the lock, will sleep and retry")
			} else {
				log.Warn().
					Err(closeErr).
					Msgf("Unexpected error attempting to close lock file: %v", closeErr)
			}
		}

		// wait for next attempt or timeout and handle appropriately
		select {
		case <-time.After(backoff):
			backoff = l.multiplyBackoff(backoff)
			attempt++
			continue
		case <-ctx.Done():
			return errors.Errorf("timed out after %s waiting for lock: %v", l.options.MaxWait, ctx.Err())
		}
	}
}

func (l *lock) multiplyBackoff(backoff time.Duration) time.Duration {
	product := float64(backoff) * l.options.BackoffMultiplier
	return time.Duration(int64(product))
}

func (l *lock) deleteExpiredLock(object object.Object, logger zerolog.Logger) error {
	logger = logger.With().Dur("expires-after", l.options.ExpiresAfter).Logger()

	logger.Debug().Msgf("Checking for expired lock")

	attrs, err := object.Handle.Attrs(object.Ctx)
	if err == storage.ErrObjectNotExist {
		log.Debug().Msgf("No existing lock found")
		return nil
	}
	if err != nil {
		return errors.Errorf("error reading attributes of lock object: %v", err)
	}

	lockAge := time.Since(attrs.Created)
	logger = logger.With().
		Time("created", attrs.Created).
		Dur("age", lockAge).
		Logger()

	if lockAge < l.options.ExpiresAfter {
		// lock file exists but is not expired
		logger.Debug().Msgf("Existing lock is not expired, won't delete it")
		return nil
	}

	logger.Debug().Msgf("Existing lock is expired, deleting it")
	// Use a generation precondition to make sure we don't run into a race condition with another process
	withCondition := object.Handle.If(storage.Conditions{GenerationMatch: attrs.Generation})

	if err = withCondition.Delete(object.Ctx); err != nil {
		if isPreconditionFailed(err) {
			logger.Warn().Msgf("Another process deleted the expired lock before we could")
			return nil
		}
		return errors.Errorf("error deleting expired lock file: %v", err)
	}

	return nil
}

// returns true if the given error is PreconditionFailed
func isPreconditionFailed(err error) bool {
	if err == nil {
		return false
	}
	if googleErr, ok := err.(*googleapi.Error); ok {
		if googleErr.Code == http.StatusPreconditionFailed {
			return true
		}
	}
	return false
}

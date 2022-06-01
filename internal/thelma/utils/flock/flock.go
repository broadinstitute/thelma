package flock

import (
	"context"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

// flock is a thin wrapper around the open-source flock library with a more user-friendly interface

type Locker interface {
	WithLock(cb func() error) error
}

// Options holds necessary attributes for a file lock with a timeout.
type Options struct {
	RetryInterval time.Duration // RetryInterval how often to retry lock attempts
	Timeout       time.Duration // Timeout how long to wait for the lock before giving up
}

type Option func(*Options)

func NewLocker(lockFile string, options ...Option) Locker {
	opts := Options{
		Timeout:       5 * time.Second,
		RetryInterval: 100 * time.Millisecond,
	}

	for _, optFn := range options {
		optFn(&opts)
	}

	return &locker{
		file:    lockFile,
		options: opts,
	}
}

type locker struct {
	file    string
	options Options
}

func (l *locker) WithLock(userFn func() error) error {
	lock, err := l.tryLock()
	if err != nil {
		return err
	}

	err = userFn()

	unlockErr := lock.Unlock()
	if err == nil {
		return unlockErr
	}
	if unlockErr != nil {
		log.Error().Err(unlockErr).Msgf("error releasing lock on %s: %v", l.file, unlockErr)
	}
	return err
}

func (l *locker) tryLock() (*flock.Flock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), l.options.Timeout)
	defer cancel()

	// Wait for lock
	log.Debug().Msgf("Attempting to acquire lock on %s, will time out after %s", l.file, l.options.Timeout)

	lock := flock.New(l.file)
	locked, err := lock.TryLockContext(ctx, l.options.RetryInterval)
	if err != nil || !locked {
		return nil, errors.WithMessagef(err, "error acquiring lock on %s (timeout %s): %v", l.file, l.options.Timeout, err)
	}

	log.Debug().Msgf("Acquired lock on %s", l.file)
	return lock, nil
}

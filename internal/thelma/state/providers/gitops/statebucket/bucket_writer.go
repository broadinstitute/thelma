package statebucket

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/lock"
	"github.com/rs/zerolog/log"
)

func newBucketWriter(bucket bucket.Bucket, cfg statebucketConfig) writer {
	return &bucketWriter{
		bucket: bucket,
		cfg:    cfg,
	}
}

type bucketWriter struct {
	bucket bucket.Bucket
	cfg    statebucketConfig
}

func (w *bucketWriter) read() (StateFile, error) {
	var result StateFile
	data, err := w.bucket.Read(w.cfg.Object)

	if err != nil {
		return result, fmt.Errorf("error reading state file: %v", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("error unmarshalling state file: %v\nContent:\n%s", err, string(data))
	}

	return result, nil
}

func (w *bucketWriter) write(state StateFile) error {
	content, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("error marshalling state file: %v", err)
	}

	return w.withLock(func() error {
		return w.bucket.Write(w.cfg.Object, content)
	})
}

func (w *bucketWriter) update(transformFn transformFn) error {
	err := w.withLock(func() error {
		return w.updateUnsafe(transformFn)
	})
	if err != nil {
		return fmt.Errorf("error updating state file: %v", err)
	}
	return nil
}

func (w *bucketWriter) updateUnsafe(transformFn transformFn) error {
	state, err := w.read()
	if err != nil {
		return err
	}

	newState, err := transformFn(state)
	if err != nil {
		return err
	}

	content, err := json.Marshal(newState)
	if err != nil {
		return fmt.Errorf("error marshalling state file: %v", err)
	}

	if err := w.bucket.Write(w.cfg.Object, content); err != nil {
		return fmt.Errorf("error writing state file: %v", err)
	}

	return nil
}

func (w *bucketWriter) withLock(fn func() error) error {
	locker := w.bucket.NewLocker(w.cfg.Lock.Object, w.cfg.Lock.MaxWait, func(options *lock.Options) {
		options.ExpiresAfter = w.cfg.Lock.ExpiresAfter
	})

	lockId, err := locker.Lock()
	if err != nil {
		return err
	}

	fnErr := fn()

	err = locker.Unlock(lockId)
	if err != nil {
		log.Error().Err(err).Msgf("error releasing lock %s: %v", w.cfg.Lock.Object, err)
	}

	// if we got a callback error, return it, else return lock release error
	if fnErr != nil {
		return fnErr
	}
	return err
}

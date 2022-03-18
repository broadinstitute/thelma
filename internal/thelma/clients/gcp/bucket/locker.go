package bucket

import (
	lock2 "github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/lock"
	"time"
)

// LockerOption used for configuring Locker options
type LockerOption func(*lock2.Options)

// Locker is distributed locking mechanism implemented over GCS.
// Every Locker is associated with an object in a GCS bucket.
type Locker interface {
	// ObjectName returns the name of the object associated with this lock
	ObjectName() string
	// Lock waits to acquire the lock, timing out after maxTime. It returns a lock id / object generation number
	// that must be passed in to Unlock
	Lock() (int64, error)
	// Unlock releases the lock.
	Unlock(lockId int64) error
}

// NewLocker returns a new Locker instance for the given object in the bucket. It accepts:
// * objectName: name of the object in the bucket to use for locking
// * maxWait: how long clients should wait to acquire the lock before giving up
// * options: optional parameters, see lock.Options for details
func (b *bucket) NewLocker(objectName string, maxWait time.Duration, options ...LockerOption) Locker {
	opts := lock2.Options{
		MaxWait:                 maxWait,
		ExpiresAfter:            0,
		BackoffMultiplier:       2,
		BackoffStartingInterval: 10 * time.Millisecond,
	}
	for _, option := range options {
		option(&opts)
	}
	return &locker{
		objectName: objectName,
		bucket:     b,
		options:    opts,
	}
}

// implements Locker interface
type locker struct {
	objectName string
	bucket     *bucket
	options    lock2.Options
}

func (l *locker) ObjectName() string {
	return l.objectName
}

func (l *locker) Lock() (int64, error) {
	op := lock2.NewLock(l.options)
	err := l.bucket.do(l.objectName, op)
	return op.Generation(), err
}

func (l *locker) Unlock(lockId int64) error {
	return l.bucket.do(l.objectName, lock2.NewUnlock(lockId))
}

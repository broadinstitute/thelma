package repo

import (
	"fmt"
	bucket2 "github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/lock"
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/object"
	"path"
	"time"
)

const ChartDir = "charts"
const indexObject = "index.yaml"

const defaultChartCacheControl = "public, max-age=300"
const defaultIndexCacheControl = "no-cache"

const defaultLockObject = ".repo.lk"
const defaultLockWaitTimeout = 2 * time.Minute
const defaultLockExpireTimeout = 5 * time.Minute

// Repo supports interactions with GCS-based Helm repositories
type Repo interface {
	// RepoURL returns the public URL of the repo
	RepoURL() string
	// IsLocked returns true if the repo is locked
	IsLocked() bool
	// Unlock unlocks the repository
	Unlock() error
	// Lock locks the repository
	Lock() error
	// UploadChart uploads a chart to the bucket
	UploadChart(fromPath string) error
	// UploadIndex uploads an index to the bucket
	UploadIndex(fromPath string) error
	// HasIndex returns true if this repo has an index object
	HasIndex() (bool, error)
	// DownloadIndex downloads an index file locally
	DownloadIndex(destPath string) error
}

type Option func(*Options)

type Options struct {
	LockWaitTimeout   time.Duration
	LockExpireTimeout time.Duration
	LockObject        string
	ChartCacheControl string
	IndexCacheControl string
}

type repo struct {
	bucket  bucket2.Bucket
	locker  bucket2.Locker
	lockId  int64
	options *Options
}

func NewRepo(bucket bucket2.Bucket, options ...Option) Repo {
	opts := &Options{
		LockWaitTimeout:   defaultLockWaitTimeout,
		LockExpireTimeout: defaultLockExpireTimeout,
		LockObject:        defaultLockObject,
		ChartCacheControl: defaultChartCacheControl,
		IndexCacheControl: defaultIndexCacheControl,
	}
	for _, option := range options {
		option(opts)
	}

	return &repo{
		bucket: bucket,
		locker: bucket.NewLocker(opts.LockObject, opts.LockWaitTimeout, func(lockOpts *lock.Options) {
			lockOpts.ExpiresAfter = opts.LockExpireTimeout
		}),
		options: opts,
	}
}

// RepoURL returns the external URL of the Helm repository
func (r *repo) RepoURL() string {
	return fmt.Sprintf("https://%s.storage.googleapis.com", r.bucket.Name())
}

// IsLocked returns true if the repo is locked
func (r *repo) IsLocked() bool {
	return r.lockId != 0
}

// Unlock unlocks the repository
func (r *repo) Unlock() error {
	if !r.IsLocked() {
		return fmt.Errorf("repo is not locked")
	}

	if err := r.locker.Unlock(r.lockId); err != nil {
		return err
	}

	r.lockId = 0

	return nil
}

// Lock locks the repository
func (r *repo) Lock() error {
	if r.IsLocked() {
		return fmt.Errorf("repo is already locked")
	}

	lockId, err := r.locker.Lock()
	if err != nil {
		return err
	}
	r.lockId = lockId

	return nil
}

// UploadChart uploads a chart package file to the correct path in the bucket
func (r *repo) UploadChart(fromPath string) error {
	objectPath := path.Join(ChartDir, path.Base(fromPath))
	return r.bucket.Upload(fromPath, objectPath, func(attrs object.AttrSet) object.AttrSet {
		return attrs.CacheControl(r.options.ChartCacheControl)
	})
}

// UploadIndex uploads an index file to correct path in the buck
func (r *repo) UploadIndex(fromPath string) error {
	return r.bucket.Upload(fromPath, indexObject, func(attrs object.AttrSet) object.AttrSet {
		return attrs.CacheControl(r.options.IndexCacheControl)
	})
}

// HasIndex returns true if this repo has an index object
func (r *repo) HasIndex() (bool, error) {
	return r.bucket.Exists(indexObject)
}

// DownloadIndex downloads the index object to given push
func (r *repo) DownloadIndex(destPath string) error {
	return r.bucket.Download(indexObject, destPath)
}

package bucket

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	object2 "github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/object"
	"github.com/broadinstitute/thelma/internal/thelma/utils/logid"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

type BucketOption func(options *BucketOptions)

// BucketOptions optional configuration for a Bucket
type BucketOptions struct {
	// Prefix is an optionally prefix to add to all object names in the bucket. Eg.
	// For bucket called "my-bucket" with a prefix of "my-prefix-",
	//    bucket.Read("foo") will read the object "gs://my-bucket/my-prefix-foo"
	Prefix string
}

// Bucket offers a simple interface for operations on GCS buckets
type Bucket interface {
	// Name returns the name of the bucket
	Name() string

	// Close closes gcs client associated with this bucket
	Close() error

	// Exists returns true if the object exists, false otherwise
	Exists(objectName string) (bool, error)

	// Upload uploads a local file to the bucket
	Upload(localPath string, objectName string, attrs ...object2.AttrSetter) error

	// Download downloads an object in the bucket to a local file
	Download(objectName string, localPath string) error

	// Read reads object contents
	Read(objectName string) ([]byte, error)

	// Write replaces object contents with given content
	Write(objectName string, content []byte, attrs ...object2.AttrSetter) error

	// Delete deletes the object from the bucket
	Delete(objectName string) error

	// Attrs returns the attributes of an object (eg. creation time, cache control)
	Attrs(objectName string) (*storage.ObjectAttrs, error)

	// Update updates the attributes  of an object (eg. cache control)
	Update(objectName string, attrs ...object2.AttrSetter) error

	// NewLocker returns a Locker instance for the given object
	NewLocker(objectName string, maxWait time.Duration, options ...LockerOption) Locker
}

// implements Bucket
type bucket struct {
	name   string // name of the GCS bucket
	prefix string // prefix to apply to all paths (used in testing)
	ctx    context.Context
	client *storage.Client
}

func NewBucket(bucketName string, options ...BucketOption) (Bucket, error) {
	return newBucket(bucketName, options...)
}

func newBucket(bucketName string, options ...BucketOption) (*bucket, error) {
	opts := BucketOptions{
		Prefix: "",
	}
	for _, optFn := range options {
		optFn(&opts)
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &bucket{
		name:   bucketName,
		ctx:    context.Background(),
		prefix: opts.Prefix,
		client: client,
	}, nil
}

func (b *bucket) Name() string {
	return b.name
}

func (b *bucket) Close() error {
	return b.client.Close()
}

func (b *bucket) Exists(objectName string) (bool, error) {
	op := object2.NewExists()
	err := b.do(objectName, op)
	return op.Exists(), err
}

func (b *bucket) Upload(localPath string, objectName string, attrs ...object2.AttrSetter) error {
	return b.do(objectName, object2.NewUpload(localPath, collateAttrs(attrs)))
}

func (b *bucket) Download(objectName string, localPath string) error {
	return b.do(objectName, object2.NewDownload(localPath))
}

func (b *bucket) Read(objectName string) ([]byte, error) {
	op := object2.NewRead()
	err := b.do(objectName, op)
	return op.Content(), err
}

func (b *bucket) Write(objectName string, content []byte, attrs ...object2.AttrSetter) error {
	_attrs := collateAttrs(attrs)
	return b.do(objectName, object2.NewWrite(content, _attrs))
}

func (b *bucket) Attrs(objectName string) (*storage.ObjectAttrs, error) {
	op := object2.NewAttrs()
	err := b.do(objectName, op)
	return op.Attrs(), err
}

func (b *bucket) Update(objectName string, attrs ...object2.AttrSetter) error {
	_attrs := collateAttrs(attrs)
	return b.do(objectName, object2.NewUpdate(_attrs))
}

func (b *bucket) Delete(objectName string) error {
	return b.do(objectName, object2.NewDelete())
}

// do executes an operation, adding useful contextual logging
func (b *bucket) do(objectName string, op object2.Operation) error {
	fullName := strings.Join([]string{b.prefix, objectName}, "")
	objectUrl := fmt.Sprintf("gs://%s/%s", b.name, fullName)

	// Build logger with context like
	// {
	//   "bucket": { "name": "my-bucket", "prefix": "" },
	//   "object": { "name": "my-object", "url": "gs://my-bucket/my-object" },
	//   "operation": { "type": "delete", "id": "fe435a" },
	// }
	ctx := log.With().
		Interface("bucket", struct {
			Name   string `json:"name"`
			Prefix string `json:"prefix"`
		}{
			Name:   b.name,
			Prefix: b.prefix,
		}).
		Interface("object", struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		}{
			Name: objectName,
			Url:  objectUrl,
		}).
		Interface("call", struct {
			Kind string `json:"kind"`
			Id   string `json:"id"`
		}{
			Kind: op.Kind(),
			Id:   logid.NewId(),
		})

	logger := ctx.Logger()

	logger.Debug().Msgf("%s %s", op.Kind(), objectUrl)
	startTime := time.Now()

	obj := object2.Object{
		Ctx:    b.ctx,
		Handle: b.client.Bucket(b.name).Object(fullName),
	}

	// run the operation
	err := op.Handler(obj, logger)

	// calculate operation duration and add to context
	duration := time.Since(startTime)
	event := logger.Debug()
	event.Dur("duration", duration)

	if err != nil {
		event.Str("status", "error")
		event.Err(err)
		returnErr := fmt.Errorf("%s failed: %v", op.Kind(), err)
		event.Msgf(returnErr.Error())
		return returnErr
	}

	event.Str("status", "ok")
	event.Msgf("%s finished in %s", op.Kind(), duration)
	return nil
}

// collate setters into an attrs object
func collateAttrs(setters []object2.AttrSetter) object2.AttrSet {
	var attrs object2.AttrSet
	for _, setter := range setters {
		attrs = setter(attrs)
	}
	return attrs
}

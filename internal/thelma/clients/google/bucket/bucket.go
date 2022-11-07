package bucket

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/object"
	"github.com/broadinstitute/thelma/internal/thelma/utils/logid"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"io"
	"path"
	"strings"
	"time"
)

const cloudConsoleBaseURL = "https://console.cloud.google.com/storage/browser"

type BucketOption func(options *BucketOptions)

// BucketOptions optional configuration for a Bucket
type BucketOptions struct {
	// Prefix is an optionally prefix to add to all object names in the bucket. Eg.
	// For bucket called "my-bucket" with a prefix of "my-prefix-",
	//    bucket.Read("foo") will read the object "gs://my-bucket/my-prefix-foo"
	Prefix string
	// ClientOptions options to pass to storage client
	ClientOptions []option.ClientOption
	// Context use a custom context instead of context.Background
	Context context.Context
}

func WithClientOptions(options ...option.ClientOption) BucketOption {
	return func(b *BucketOptions) {
		b.ClientOptions = append(b.ClientOptions, options...)
	}
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
	Upload(localPath string, objectName string, attrs ...object.AttrSetter) error

	// Download downloads an object in the bucket to a local file
	Download(objectName string, localPath string) error

	// Read reads object contents
	Read(objectName string) ([]byte, error)

	// Write replaces object contents with given content
	Write(objectName string, content []byte, attrs ...object.AttrSetter) error

	// WriteFromStream replaces object contents with given content
	WriteFromStream(objectName string, reader io.Reader, attrs ...object.AttrSetter) error

	// Delete deletes the object from the bucket
	Delete(objectName string) error

	// Attrs returns the attributes of an object (eg. creation time, cache control)
	Attrs(objectName string) (*storage.ObjectAttrs, error)

	// Update updates the attributes  of an object (eg. cache control)
	Update(objectName string, attrs ...object.AttrSetter) error

	// NewLocker returns a Locker instance for the given object
	NewLocker(objectName string, maxWait time.Duration, options ...LockerOption) Locker

	// CloudConsoleURL returns a URL pointing an object in the Google cloud console
	// See https://cloud.google.com/storage/docs/request-endpoints#console
	// Returns https://console.cloud.google.com/storage/browser/<BUCKET_NAME>/<OBJECT_NAME>
	CloudConsoleURL(objectNameOrPrefix string) string
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
		Prefix:  "",
		Context: context.Background(),
	}
	for _, optFn := range options {
		optFn(&opts)
	}

	client, err := storage.NewClient(opts.Context, opts.ClientOptions...)
	if err != nil {
		return nil, err
	}

	return &bucket{
		name:   bucketName,
		ctx:    opts.Context,
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
	op := object.NewExists()
	err := b.do(objectName, op)
	return op.Exists(), err
}

func (b *bucket) Upload(localPath string, objectName string, attrs ...object.AttrSetter) error {
	return b.do(objectName, object.NewUpload(localPath, collateAttrs(attrs)))
}

func (b *bucket) Download(objectName string, localPath string) error {
	return b.do(objectName, object.NewDownload(localPath))
}

func (b *bucket) Read(objectName string) ([]byte, error) {
	op := object.NewRead()
	err := b.do(objectName, op)
	return op.Content(), err
}

func (b *bucket) Write(objectName string, content []byte, attrs ...object.AttrSetter) error {
	_attrs := collateAttrs(attrs)
	return b.do(objectName, object.NewWrite(content, _attrs))
}

func (b *bucket) WriteFromStream(objectName string, reader io.Reader, attrs ...object.AttrSetter) error {
	_attrs := collateAttrs(attrs)
	return b.do(objectName, object.NewWriteFromStream(reader, _attrs))
}

func (b *bucket) Attrs(objectName string) (*storage.ObjectAttrs, error) {
	op := object.NewAttrs()
	err := b.do(objectName, op)
	return op.Attrs(), err
}

func (b *bucket) Update(objectName string, attrs ...object.AttrSetter) error {
	_attrs := collateAttrs(attrs)
	return b.do(objectName, object.NewUpdate(_attrs))
}

func (b *bucket) Delete(objectName string) error {
	return b.do(objectName, object.NewDelete())
}

func (b *bucket) CloudConsoleURL(objectNameOrPrefix string) string {
	return CloudConsoleURL(b.name, path.Join(b.prefix, objectNameOrPrefix))
}

func CloudConsoleURL(bucketName string, objectNameOrPrefix string) string {
	return fmt.Sprintf("%s/%s/%s", cloudConsoleBaseURL, bucketName, objectNameOrPrefix)
}

// do executes an operation, adding useful contextual logging
func (b *bucket) do(objectName string, op object.Operation) error {
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

	logger.Trace().Msgf("%s %s", op.Kind(), objectUrl)
	startTime := time.Now()

	obj := object.Object{
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
func collateAttrs(setters []object.AttrSetter) object.AttrSet {
	var attrs object.AttrSet
	for _, setter := range setters {
		attrs = setter(attrs)
	}
	return attrs
}

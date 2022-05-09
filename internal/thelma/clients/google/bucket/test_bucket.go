//go:build smoke
// +build smoke

package bucket

// This file uses the `smoke` build tag so that that only tests that also have the `smoke` build tag can use it.

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket/object"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// Bucket for testing Thelma code that interacts with GCS; lives in dsp-tools-k8s project.
const testBucketName = "thelma-integration-tests"

type testBucket struct {
	*bucket
	tracker *objectTracker
}

// any objects written during the test are tracked here so they can be automatically cleaned up
func newObjectTracker() *objectTracker {
	return &objectTracker{
		objectsToCleanup: set.NewStringSet(),
	}
}

type objectTracker struct {
	mutex            sync.Mutex
	objectsToCleanup set.StringSet
}

func (t *objectTracker) add(objectName string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.objectsToCleanup.Add(objectName)
}

// NewTestBucket (FOR USE IN TESTS ONLY) creates a Bucket for use in
// integration tests.
//
// It writes objects to Thelma's integration test bucket, adding
// a random prefix to all object paths. This is so the same
// test can be safely executed concurrently.
//
// Example:
// func MyTest(t *testing.T) {
//   b := NewTestBucket("my-integration-test")
//
//   // do things with the bucket, eg.
//	 assert.NoError(t, b.Write("my-file", []byte("data")))
// }
//
func NewTestBucket(t *testing.T) Bucket {
	// generate prefix for this test bucket instance
	prefix := testPrefix(t.Name())
	log.Debug().Msgf("test prefix for %s: %q", t.Name(), prefix)

	// create a new underlying bucket instance
	_bucket, err := newBucket(testBucketName, func(options *BucketOptions) {
		options.Prefix = prefix
	})
	require.NoError(t, err, "failed to initialize bucket")

	// wrap it in a test bucket
	_testBucket := &testBucket{
		bucket:  _bucket,
		tracker: newObjectTracker(),
	}

	// add a cleanup function to delete any objects written during the test
	t.Cleanup(func() {
		require.NoError(t, _testBucket.Close(), "failed to close bucket")
	})

	return _testBucket
}

// Below we override any bucket functions that can write objects to the bucket

func (b *testBucket) Upload(localPath string, objectName string, attrs ...object.AttrSetter) error {
	b.tracker.add(objectName)
	return b.bucket.Upload(localPath, objectName, attrs...)
}

func (b *testBucket) Write(objectName string, content []byte, attrs ...object.AttrSetter) error {
	b.tracker.add(objectName)
	return b.bucket.Write(objectName, content, attrs...)
}

func (b *testBucket) NewLocker(objectName string, maxWait time.Duration, options ...LockerOption) Locker {
	b.tracker.add(objectName)
	return b.bucket.NewLocker(objectName, maxWait, options...)
}

// Close deletes any remaining test objects in the bucket and closes the gcs client.
func (b *testBucket) Close() error {
	log.Debug().Msgf("Cleaning up all objects in bucket %s with prefix: %s", b.name, b.prefix)

	// note that we track all objects that are written.
	// We can't use GCS's List support because it's eventually consistent,
	// so objects written during the test might not yet show up in List calls when this function is run
	for _, objectName := range b.tracker.objectsToCleanup.Elements() {
		if err := b.Delete(objectName); err != nil {
			if strings.Contains(err.Error(), "storage: object doesn't exist") {
				// TODO return underlying error from GCS so we can check its type instead of error message content
				log.Debug().Msgf("couldn't delete %s, likely already deleted by test: %v", objectName, err)
				// test likely deleted the object already
			} else {
				return fmt.Errorf("error cleaning up test object %s: %v", objectName, err)
			}
		}
	}

	return b.bucket.Close()
}

// testPrefix generates a new random prefix for the test bucket in the form
// "my-unit-test/20060102.150405/4d658221
func testPrefix(identifier string) string {
	timestamp := time.Now().UTC().Format("20060102.150405")
	return fmt.Sprintf("%s/%s/%08x/", identifier, timestamp, rand.Int31())
}

//go:build smoke
// +build smoke

package bucket

// This file uses the `smoke` build tag so that that only tests that also have the `smoke` build tag can use it.

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/gcp/bucket/object"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// Bucket for testing Thelma code that interacts with GCS; lives in dsp-tools-k8s project
const testBucketName = "thelma-gcs-integration-test"

type testBucket struct {
	*bucket
	mutex            sync.Mutex
	objectsToCleanup []string // any objects written during the test are tracked here so they can be automatically cleaned up
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
		bucket: _bucket,
	}

	// add a cleanup function to delete any objects written during the test
	t.Cleanup(func() {
		require.NoError(t, _testBucket.Close(), "failed to close bucket")
	})

	return _testBucket
}

func (b *testBucket) Upload(localPath string, objectName string, attrs ...object.AttrSetter) error {
	b.addToCleanup(objectName)
	return b.bucket.Upload(localPath, objectName, attrs...)
}

func (b *testBucket) Write(objectName string, content []byte, attrs ...object.AttrSetter) error {
	b.addToCleanup(objectName)
	return b.bucket.Write(objectName, content, attrs...)
}

func (b *testBucket) addToCleanup(objectName string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.objectsToCleanup = append(b.objectsToCleanup, objectName)
}

// Close deletes any remaining test objects in the bucket and closes the gcs client.
func (b *testBucket) Close() error {
	log.Debug().Msgf("Cleaning up all objects in bucket %s with prefix: %s", b.name, b.prefix)

	// note that we track all objects that are written.
	// We can't use GCS's List support because it's eventually consistent,
	// so objects written during the test might not yet show up in List calls when this function is run
	for _, objectName := range b.objectsToCleanup {
		if err := b.Delete(objectName); err != nil {
			return fmt.Errorf("error cleaning up test object %s: %v", objectName, err)
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

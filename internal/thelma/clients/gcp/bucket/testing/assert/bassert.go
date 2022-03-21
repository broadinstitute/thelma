// Package assert contains helper functions for making Testify assertions about objects in GCS buckets.
package assert

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ObjectExists(t *testing.T, bucket bucket.Bucket, objectName string, msgAndArgs ...interface{}) bool {
	exists, err := bucket.Exists(objectName)
	if !assert.NoError(t, err, "unexpected error checking existence of object %s in bucket", objectName) {
		return false
	}
	return assert.True(t, exists, withMessage("object %s should exist in bucket, but does not", objectName).add(msgAndArgs)...)
}

func NoObjectExists(t *testing.T, bucket bucket.Bucket, objectName string, msgAndArgs ...interface{}) bool {
	exists, err := bucket.Exists(objectName)
	if !assert.NoError(t, err, "unexpected error checking existence of object %s in bucket", objectName) {
		return false
	}
	return assert.False(t, exists, withMessage("object %s should not exist in bucket, but does", objectName).add(msgAndArgs)...)
}

func ObjectHasContent(t *testing.T, bucket bucket.Bucket, objectName string, content string, msgAndArgs ...interface{}) bool {
	actual, err := bucket.Read(objectName)
	if !assert.NoError(t, err, "error reading object %s: %v", objectName, err) {
		return false
	}
	return assert.Equal(t, content, string(actual), msgAndArgs...)
}

func ObjectHasCacheControl(t *testing.T, bucket bucket.Bucket, objectName string, cacheControl string, msgAndArgs ...interface{}) bool {
	attrs, err := bucket.Attrs(objectName)
	if !assert.NoError(t, err, "error reading attributes of object %s: %v", objectName, err) {
		return false
	}
	return assert.Equal(t, cacheControl, attrs.CacheControl, withMessage("expected cache control attribute for object %s to be %q", objectName, cacheControl).add(msgAndArgs)...)
}

func ObjectHasGeneration(t *testing.T, bucket bucket.Bucket, objectName string, generation int64, msgAndArgs ...interface{}) bool {
	attrs, err := bucket.Attrs(objectName)
	if !assert.NoError(t, err, "error reading attributes of object %s: %v", objectName, err) {
		return false
	}
	return assert.Equal(t, generation, attrs.Generation, withMessage("expected generation for object %s to be %d", objectName, generation).add(msgAndArgs)...)
}

// private sugar for constructing messages
type msg struct {
	prefix []interface{}
}

func withMessage(msgAndArgs ...interface{}) msg {
	return msg{
		prefix: msgAndArgs,
	}
}

func headAndTail(msgAndArgs []interface{}) (string, []interface{}) {
	format, ok := msgAndArgs[0].(string)
	if !ok {
		panic(fmt.Errorf("first argument should be format string: %v", msgAndArgs))
	}
	return format, msgAndArgs[1:]
}

func (m msg) add(msgAndArgs []interface{}) []interface{} {
	if len(msgAndArgs) == 0 {
		return m.prefix
	}
	if len(m.prefix) == 0 {
		return msgAndArgs
	}
	pfmt, pargs := headAndTail(m.prefix)
	mfmt, margs := headAndTail(msgAndArgs)

	prefix := fmt.Sprintf(pfmt, pargs...)
	combinedfmt := fmt.Sprintf("%s: %s", prefix, mfmt)

	var result []interface{}
	result = append(result, combinedfmt)
	result = append(result, margs...)
	return result
}

package require

import (
	"github.com/broadinstitute/thelma/internal/thelma/utils/gcp/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/utils/gcp/bucket/testing/assert"
	"testing"
)

// brequire: wrapper around `bassert` functions that calls t.FailNow() if the assertion fails (works like testify's require package)

func ObjectExists(t *testing.T, bucket bucket.Bucket, objectName string, msgAndArgs ...interface{}) {
	if !assert.ObjectExists(t, bucket, objectName, msgAndArgs...) {
		t.FailNow()
	}
}

func NoObjectExists(t *testing.T, bucket bucket.Bucket, objectName string, msgAndArgs ...interface{}) {
	if !assert.NoObjectExists(t, bucket, objectName, msgAndArgs...) {
		t.FailNow()
	}
}

func ObjectHasContent(t *testing.T, bucket bucket.Bucket, objectName string, content string, msgAndArgs ...interface{}) {
	if !assert.ObjectHasContent(t, bucket, objectName, content, msgAndArgs...) {
		t.FailNow()
	}
}

func ObjectHasCacheControl(t *testing.T, bucket bucket.Bucket, objectName string, cacheControl string, msgAndArgs ...interface{}) {
	if !assert.ObjectHasCacheControl(t, bucket, objectName, cacheControl, msgAndArgs...) {
		t.FailNow()
	}
}

func ObjectHasGeneration(t *testing.T, bucket bucket.Bucket, objectName string, generation int64, msgAndArgs ...interface{}) {
	if !assert.ObjectHasGeneration(t, bucket, objectName, generation, msgAndArgs...) {
		t.FailNow()
	}
}

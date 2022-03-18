//go:build smoke
// +build smoke

package tests

// This file contains a simple example smoke test demonstrating the use of TestBucket

import (
	bucket2 "github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket"
	bassert "github.com/broadinstitute/thelma/internal/thelma/clients/gcp/bucket/testing/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// HelloBucket is a simple function that writes a file to a GCS bucket, returning an error if the write fails
func HelloBucket(b bucket2.Bucket) error {
	return b.Write("hello.txt", []byte("hello, world"))
}

// TestHelloBucket tests the HelloBucket function
func TestHelloBucket(t *testing.T) {
	b := bucket2.NewTestBucket(t)
	err := HelloBucket(b)
	require.NoError(t, err)
	bassert.ObjectExists(t, b, "hello.txt")
	bassert.ObjectHasContent(t, b, "hello.txt", "hello, world")
}

package tests

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CloudConsoleURL(t *testing.T) {
	b, err := bucket.NewBucket("broad-dsp-fake-bucket-does-not-exist-12038493asdfhasdhfhsadkjhlhl", func(options *bucket.BucketOptions) {
		options.Prefix = "my/prefix"
	})
	require.NoError(t, err)
	assert.Equal(t, "https://console.cloud.google.com/storage/browser/broad-dsp-fake-bucket-does-not-exist-12038493asdfhasdhfhsadkjhlhl/my/prefix/my/object", b.CloudConsoleURL("my/object"))

	// test package-level function as well
	assert.Equal(t, "https://console.cloud.google.com/storage/browser/fake-bucket/a/b/object", bucket.CloudConsoleURL("fake-bucket", "a/b/object"))
}

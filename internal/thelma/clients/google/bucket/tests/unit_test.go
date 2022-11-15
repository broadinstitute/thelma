package tests

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CloudConsoleURLs(t *testing.T) {
	assert.Equal(t, "https://console.cloud.google.com/storage/browser/fake-bucket/a/b", bucket.CloudConsoleObjectListURL("fake-bucket", "a/b"))
	assert.Equal(t, "https://console.cloud.google.com/storage/browser/_details/fake-bucket/a/b/object", bucket.CloudConsoleObjectDetailURL("fake-bucket", "a/b/object"))
}

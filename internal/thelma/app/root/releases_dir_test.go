package root

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestReleasesDir(t *testing.T) {
	dir := t.TempDir()
	_root := New(dir)
	_releasesDir := _root.ReleasesDir()
	assert.Equal(t, path.Join(dir, "releases"), _releasesDir.Root())
	assert.Equal(t, path.Join(dir, "releases", "current"), _releasesDir.CurrentSymlink())
	assert.Equal(t, path.Join(dir, "releases", version.Version), _releasesDir.ForCurrentVersion())
	assert.Equal(t, path.Join(dir, "releases", "blah"), _releasesDir.ForVersion("blah"))
}

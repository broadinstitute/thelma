package toolbox

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_Toolbox__ExpandPath(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(path.Join(dir, "foo"), []byte("fake tool"), 0755))

	_toolbox, err := NewToolFinderWithDir(dir)
	require.NoError(t, err)

	assert.Equal(t, path.Join(dir, "foo"), _toolbox.ExpandPath("foo"), "foo path should be fully qualified")
	assert.Equal(t, "bar", _toolbox.ExpandPath("bar"), "bar should not be qualified since it does not exist in tool dir")
}

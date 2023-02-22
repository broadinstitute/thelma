package manifest

import (
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_Manifest(t *testing.T) {
	dir := t.TempDir()
	err := EnsureMatches(dir, "v1.2.3")
	require.Error(t, err)

	err = os.WriteFile(path.Join(dir, filename), []byte(`{"version":"v1.2.3"}`), 0600)
	require.NoError(t, err)

	require.NoError(t, EnsureMatches(dir, "v1.2.3"))
}

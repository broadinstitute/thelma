package stores

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path"
	"testing"
)

func Test_DirectoryStore(t *testing.T) {
	s, err := NewDirectoryStore(path.Join(t.TempDir(), "does-not-exist"))
	require.NoError(t, err)

	exists, err := s.Exists("my-key")
	require.NoError(t, err)
	assert.False(t, exists)

	err = s.Write("my-key", []byte("super secret"))
	require.NoError(t, err)

	exists, err = s.Exists("my-key")
	require.NoError(t, err)
	assert.True(t, exists)

	credential, err := s.Read("my-key")
	require.NoError(t, err)
	assert.Equal(t, "super secret", string(credential))
}

func Test_NoopStore(t *testing.T) {
	s := NewNoopStore()

	exists, err := s.Exists("my-key")
	require.NoError(t, err)
	assert.False(t, exists)

	err = s.Write("my-key", []byte("super secret"))
	require.NoError(t, err)

	exists, err = s.Exists("my-key")
	require.NoError(t, err)
	assert.False(t, exists)

	credential, err := s.Read("my-key")
	require.NoError(t, err)
	assert.Empty(t, credential)
}

package vault

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_VaultTokenStore(t *testing.T) {
	fakeToken := []byte("fake-vault-token")
	homeDir := t.TempDir()
	s := newVaultTokenStore(homeDir)

	exists, err := s.Exists("ignored")
	require.NoError(t, err)
	assert.False(t, exists)

	err = s.Write("ignored", fakeToken)
	require.NoError(t, err)

	require.FileExists(t, path.Join(homeDir, ".vault-token"))
	content, err := os.ReadFile(path.Join(homeDir, ".vault-token"))
	require.NoError(t, err)
	assert.Equal(t, string(fakeToken), string(content))

	exists, err = s.Exists("ignored")
	require.NoError(t, err)
	assert.True(t, exists)

	credential, err := s.Read("my-key")
	require.NoError(t, err)
	assert.Equal(t, string(fakeToken), string(credential))

	err = s.Remove("my-key")
	require.NoError(t, err)
}

package vault

import (
	vaulthelper "github.com/broadinstitute/thelma/internal/thelma/clients/vault/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_MaskingRoundTripper(t *testing.T) {
	fakeVaultServer := vaulthelper.NewFakeVaultServer(t)
	fakeVaultServer.SetSecret("secret/foo/bar", map[string]interface{}{"abc": "xyz"})

	client := fakeVaultServer.Server().Client()

	maskFnCalled := false

	maskingTransport := MaskingRoundTripper{
		inner: client.Transport,
		maskFn: func(secrets ...string) {
			assert.Len(t, secrets, 1)
			assert.Equal(t, "xyz", secrets[0])
			maskFnCalled = true
		},
	}
	client.Transport = maskingTransport

	_, err := client.Get(fakeVaultServer.Server().URL + "/v1/secret/foo/bar")
	require.NoError(t, err)
	assert.True(t, maskFnCalled)
}

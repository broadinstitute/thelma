package vault

import (
	vaulthelper "github.com/broadinstitute/thelma/internal/thelma/clients/vault/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_MaskingRoundTripper(t *testing.T) {
	// Set up a fake Vault http server
	fakeVaultServer := vaulthelper.NewFakeVaultServer(t)
	// Add a fake secret to the vault server -- our test will retrieve it
	fakeVaultServer.SetSecret("secret/foo/bar", map[string]interface{}{"abc": "xyz"})

	// Get an HTTP client configured to talk to the fake Vault http server
	client := fakeVaultServer.Server().Client()

	// By default, the round tripper calls logging.MaskSeret to mask secrets.
	// We supply a custom fake masking function here, so we can verify we're calling with the right parameters.
	maskFnCalled := false
	maskFn := func(secrets ...string) {
		assert.Len(t, secrets, 1)
		assert.Equal(t, "xyz", secrets[0])
		maskFnCalled = true
	}

	// create a new MaskingRoundTripper that wraps the client's default transport
	maskingTransport := newMaskingRoundTripper(client.Transport)
	// configure it to use our custom masking function
	maskingTransport.maskFn = maskFn
	client.Transport = maskingTransport

	// And now, the test!
	// Let's retrieve the secret
	_, err := client.Get(fakeVaultServer.Server().URL + "/v1/secret/foo/bar")

	// Make sure there were no errors
	require.NoError(t, err)

	// Make sure our fake masking function was called
	assert.True(t, maskFnCalled)
}

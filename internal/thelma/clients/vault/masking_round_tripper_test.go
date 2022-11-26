package vault

import (
	vaulthelper "github.com/broadinstitute/thelma/internal/thelma/clients/vault/testing"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_MaskingRoundTripper(t *testing.T) {
	// Set up a fake Vault http server
	fakeVaultServer := vaulthelper.NewFakeVaultServer(t)
	// Add a fake secret to the vault server -- our test will retrieve it
	fakeVaultServer.SetSecret("secret/foo/bar",
		map[string]interface{}{
			"field1":        "mask-me-1",
			"field2":        "mask-me-2",
			"project_id":    "should-not-be-masked-1",
			"user":          "should-not-be-masked-2",
			"too-short":     "1234",  // should not be masked
			"not-too-short": "12345", // should be masked
		},
	)

	// Get an HTTP client configured to talk to the fake Vault http server
	client := fakeVaultServer.Server().Client()

	// By default, the round tripper calls logging.MaskSeret to mask secrets.
	// We supply a custom fake masking function here, so we can verify we're calling with the right parameters.
	maskedValues := set.NewStringSet()
	maskFn := func(secrets ...string) {
		assert.Len(t, secrets, 1)
		maskedValues.Add(secrets...)
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

	// Make sure the expected values were masked
	assert.Equal(t, 3, maskedValues.Size())
	assert.ElementsMatch(t, []string{"mask-me-1", "mask-me-2", "12345"}, maskedValues.Elements())
}

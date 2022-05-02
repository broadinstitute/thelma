package vault

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	credshelper "github.com/broadinstitute/thelma/internal/thelma/app/credentials/testing"
	vaulthelper "github.com/broadinstitute/thelma/internal/thelma/clients/vault/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const fakeGithubToken = "my-github-token"

func Test_TokenProvider(t *testing.T) {
	// create fake vault server
	fakeVaultServer := vaulthelper.NewFakeVaultServer(t)
	fakeVaultServer.ExpectGithubLogin(fakeGithubToken, "my-new-vault-token")

	// use an empty token store instead of the default store, which points at ~/.vault-token
	tokenStore, err := stores.NewDirectoryStore(t.TempDir())
	require.NoError(t, err)

	thelmaConfig, err := config.Load(config.WithTestDefaults(t))
	require.NoError(t, err)

	// create fake credential store
	creds, err := credshelper.NewFakeCredentials(t)
	require.NoError(t, err)
	err = creds.AddToStore(githubTokenCredentialKey, []byte(fakeGithubToken))
	require.NoError(t, err)

	// make a new token provider configured to use fake vault server and token store
	provider, err := TokenProvider(thelmaConfig, creds, func(options *ClientOptions) {
		options.CredentialStore = tokenStore
		options.ConfigureVaultClient(fakeVaultServer.ConfigureClient)
	})
	require.NoError(t, err)

	token, err := provider.Get()
	require.NoError(t, err)

	assert.Equal(t, "my-new-vault-token", string(token))
}

package vault

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	vaultapi "github.com/hashicorp/vault/api"
	"os"
)

const configKey = "vault"
const githubTokenCredentialKey = "github-token"
const vaultTokenCredentialKey = "vault-token"

const githubLoginPath = "/auth/github/login"
const tokenLookupPath = "/auth/token/lookup"

type vaultConfig struct {
	Addr string `default:"https://clotho.broadinstitute.org:8200"`
}

type ClientOptions struct {
	CredentialStore    stores.Store
	VaultClientOptions []func(*vaultapi.Config)
}

func (c *ClientOptions) ConfigureVaultClient(configureFn func(*vaultapi.Config)) {
	c.VaultClientOptions = append(c.VaultClientOptions, configureFn)
}

type ClientOption func(*ClientOptions)

// NewClient returns a new authenticated vault client
func NewClient(thelmaConfig config.Config, creds credentials.Credentials, clientOptions ...ClientOption) (*vaultapi.Client, error) {
	options := aggregateOptions(clientOptions...)

	client, err := newUnauthenticatedClient(thelmaConfig, options)
	if err != nil {
		return nil, err
	}

	provider := buildVaultTokenProvider(client, creds, options)
	tokenValue, err := provider.Get()
	if err != nil {
		return nil, err
	}
	client.SetToken(string(tokenValue))
	return client, nil
}

// TokenProvider returns a new credentials.TokenProvider that provides a Vault token
func TokenProvider(thelmaConfig config.Config, creds credentials.Credentials, clientOptions ...ClientOption) (credentials.TokenProvider, error) {
	options := aggregateOptions(clientOptions...)

	client, err := newUnauthenticatedClient(thelmaConfig, options)
	if err != nil {
		return nil, err
	}

	return buildVaultTokenProvider(client, creds, options), nil
}

func aggregateOptions(clientOptions ...ClientOption) *ClientOptions {
	var opts ClientOptions

	for _, clientOption := range clientOptions {
		clientOption(&opts)
	}

	if opts.CredentialStore == nil {
		opts.CredentialStore = NewVaultTokenStore()
	}

	return &opts
}

func buildVaultTokenProvider(unauthedClient *vaultapi.Client, creds credentials.Credentials, opts *ClientOptions) credentials.TokenProvider {
	githubToken := buildGithubTokenProvider(unauthedClient, creds)

	return creds.NewTokenProvider(vaultTokenCredentialKey, func(options *credentials.TokenOptions) {
		// Use custom credential store that stores token at ~/.vault-token instead ~/.thelma/credentials
		options.CredentialStore = opts.CredentialStore

		options.IssueFn = func() ([]byte, error) {
			githubPAT, err := githubToken.Get()
			if err != nil {
				return nil, fmt.Errorf("could not issue new Vault token, failed to load Github PAT: %v", err)
			}
			vaultToken, err := login(unauthedClient, string(githubPAT))
			if err != nil {
				return nil, fmt.Errorf("failed to issue new Vault token: %v", err)
			}
			return []byte(vaultToken), nil
		}

		// The Vault token is valid if it can be used to make a token lookup call
		options.ValidateFn = func(vaultToken []byte) error {
			return tokenLookup(unauthedClient, string(vaultToken))
		}
	})
}

func buildGithubTokenProvider(unauthedClient *vaultapi.Client, creds credentials.Credentials) credentials.TokenProvider {
	return creds.NewTokenProvider(githubTokenCredentialKey, func(options *credentials.TokenOptions) {
		options.PromptEnabled = true

		options.PromptMessage = `
A GitHub Personal Access Token is required to authenticate to Vault.
You can generate a new PAT at https://github.com/settings/tokens (select ONLY the read:org scope).

Enter Personal Access Token: `

		// The GitHub PAT is valid if it can be used to authenticate to Vault
		options.ValidateFn = func(githubPat []byte) error {
			_, err := login(unauthedClient, string(githubPat))
			return err
		}
	})
}

// newUnauthenticatedClient returns a new unauthenticated vault client
func newUnauthenticatedClient(thelmaConfig config.Config, options *ClientOptions) (*vaultapi.Client, error) {
	vaultCfg, err := loadConfig(thelmaConfig)
	if err != nil {
		return nil, err
	}

	clientCfg := vaultapi.DefaultConfig() // modify for more granular configuration
	clientCfg.Address = vaultCfg.Addr

	for _, optionFn := range options.VaultClientOptions {
		optionFn(clientCfg)
	}

	// wrap default transport in a MaskingRoundTripper, which automatically masks values in Secrets
	transport := newMaskingRoundTripper(clientCfg.HttpClient.Transport)
	clientCfg.HttpClient.Transport = transport

	return vaultapi.NewClient(clientCfg)
}

func loadConfig(thelmaConfig config.Config) (vaultConfig, error) {
	var cfg vaultConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Addr == "" {
		cfg.Addr = os.Getenv("VAULT_ADDR")
	}
	return cfg, nil
}

// login performs a login API request (the equivalent of "vault login -method=github" on the command-line)
// https://www.vaultproject.io/api-docs/auth/github#login
func login(client *vaultapi.Client, githubToken string) (string, error) {
	_client, err := client.Clone()
	if err != nil {
		return "", err
	}

	secret, err := _client.Logical().Write(githubLoginPath, map[string]interface{}{"token": githubToken})
	if err != nil {
		return "", fmt.Errorf("login request failed: %v", err)
	}
	return secret.Auth.ClientToken, nil
}

// tokenLookup performs a token lookup API request (the equivalent of "vault token lookup" on the command-line)
// https://www.vaultproject.io/api-docs/auth/token#lookup-a-token
func tokenLookup(client *vaultapi.Client, vaultToken string) error {
	_client, err := client.Clone()
	if err != nil {
		return err
	}

	_client.SetToken(vaultToken)

	// we don't actually care about any data in the response, just that the token lookup succeeds
	_, err = _client.Logical().Write(tokenLookupPath, nil)

	return err
}

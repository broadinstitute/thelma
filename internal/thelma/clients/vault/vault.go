package vault

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"os"
)

const configKey = "vault"
const githubTokenCredentialKey = "github-token"
const vaultTokenCredentialKey = "vault-token"

const githubLoginPath = "/auth/github/login"
const approleLoginPath = "/auth/approle/login"
const tokenLookupPath = "/auth/token/lookup-self"

const approleRoleIdEnvVar = "VAULT_ROLE_ID"
const approleSecretIdEnvVar = "VAULT_SECRET_ID"

type vaultConfig struct {
	Addr            string `default:"https://clotho.broadinstitute.org:8200"`
	ManageUserToken bool   `default:"true"`
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
	options, err := buildClientOptions(thelmaConfig, clientOptions...)
	if err != nil {
		return nil, err
	}

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
	options, err := buildClientOptions(thelmaConfig, clientOptions...)
	if err != nil {
		return nil, err
	}

	client, err := newUnauthenticatedClient(thelmaConfig, options)
	if err != nil {
		return nil, err
	}

	return buildVaultTokenProvider(client, creds, options), nil
}

func buildClientOptions(thelmaConfig config.Config, clientOptions ...ClientOption) (*ClientOptions, error) {
	vaultCfg, err := loadConfig(thelmaConfig)
	if err != nil {
		return nil, err
	}

	options := new(ClientOptions)

	if vaultCfg.ManageUserToken {
		// if we're managing the user vault token (~/.vault-token), pass in our custom token store.
		options.CredentialStore = NewVaultTokenStore()
	}

	for _, clientOption := range clientOptions {
		clientOption(options)
	}

	return options, nil
}

func buildVaultTokenProvider(unauthedClient *vaultapi.Client, creds credentials.Credentials, opts *ClientOptions) credentials.TokenProvider {
	githubToken := buildGithubTokenProvider(unauthedClient, creds)

	return creds.NewTokenProvider(vaultTokenCredentialKey, func(options *credentials.TokenOptions) {
		// Use custom credential store that stores token at ~/.vault-token instead ~/.thelma/credentials
		options.CredentialStore = opts.CredentialStore

		options.IssueFn = func() ([]byte, error) {
			githubPAT, err := githubToken.Get()
			if err != nil {
				// Couldn't get GitHub auth, see if we can log in via approle before we error out:
				approleRoleId := os.Getenv(approleRoleIdEnvVar)
				approleSecretId := os.Getenv(approleSecretIdEnvVar)
				if approleRoleId != "" && approleSecretId != "" {
					vaultToken, approleErr := approleLogin(unauthedClient, approleRoleId, approleSecretId)
					if approleErr != nil {
						// If that failed on its own merits, note that error but continue to return the primary Github error
						log.Warn().Msgf("tried to login to Vault with approle because Github failed but encountered another error: %v", approleErr)
					} else {
						// If approle worked, return it immediately so we ignore the Github error
						return []byte(vaultToken), nil
					}
				}

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

// approleLogin is like login except via role+secret ID rather than via GitHub PAT
// https://developer.hashicorp.com/vault/docs/auth/approle
func approleLogin(client *vaultapi.Client, roleId string, secretId string) (string, error) {
	_client, err := client.Clone()
	if err != nil {
		return "", err
	}

	secret, err := _client.Logical().Write(approleLoginPath, map[string]interface{}{
		"role_id":   roleId,
		"secret_id": secretId,
	})
	if err != nil {
		return "", fmt.Errorf("approle login request failed: %v", err)
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
	_, err = _client.Logical().Read(tokenLookupPath)

	return err
}

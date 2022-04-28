package vault

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	vaultapi "github.com/hashicorp/vault/api"
	"os"
)

const configKey = "vault"
const githubTokenCredentialKey = "github-token"
const vaultTokenCredentialKey = "vault-token"

type vaultConfig struct {
	Addr string `default:"https://clotho.broadinstitute.org:8200"`
}

// NewClient returns a new authenticated vault client
func NewClient(thelmaConfig config.Config, creds credentials.Credentials) (*vaultapi.Client, error) {
	vaultCfg, err := loadConfig(thelmaConfig)
	if err != nil {
		return nil, err
	}

	client, err := newUnauthenticatedClient(vaultCfg)
	if err != nil {
		return nil, err
	}

	githubToken := newGitHubToken(client, creds)
	vaultToken := newVaultToken(client, creds, githubToken)
	tokenVal, err := vaultToken.Get()
	if err != nil {
		return nil, err
	}

	client.SetToken(string(tokenVal))
	return client, nil
}

func newVaultToken(unauthedClient *vaultapi.Client, creds credentials.Credentials, githubToken credentials.Token) credentials.Token {
	return creds.NewToken(vaultTokenCredentialKey, func(options *credentials.TokenOptions) {
		// Use special credential store that stores token at ~/.vault-token instead ~/.thelma/credentials
		options.CredentialStore = NewVaultTokenStore()

		options.IssueFn = func() ([]byte, error) {
			githbuPAT, err := githubToken.Get()
			if err != nil {
				return nil, fmt.Errorf("could not issue new Vault token, failed to load Github PAT: %v", err)
			}
			vaultToken, err := login(unauthedClient, string(githbuPAT))
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

func newGitHubToken(unauthedClient *vaultapi.Client, creds credentials.Credentials) credentials.Token {
	return creds.NewToken(githubTokenCredentialKey, func(options *credentials.TokenOptions) {
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

// newUnauthenticatedClient returns a new unauthenticated vault client
func newUnauthenticatedClient(cfg vaultConfig) (*vaultapi.Client, error) {
	clientCfg := vaultapi.DefaultConfig() // modify for more granular configuration
	clientCfg.Address = cfg.Addr
	return vaultapi.NewClient(clientCfg)
}

// login performs a login API request (the equivalent of "vault login -method=github" on the command-line)
// https://www.vaultproject.io/api-docs/auth/github#login
func login(client *vaultapi.Client, githubToken string) (string, error) {
	_client, err := client.Clone()
	if err != nil {
		return "", err
	}

	secret, err := _client.Logical().Write("/auth/github/login", map[string]interface{}{"token": githubToken})
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
	_, err = _client.Logical().Write("/auth/token/lookup", nil)

	return err
}

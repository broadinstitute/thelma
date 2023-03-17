package github

import (
	"context"
	"errors"
	"net/http"

	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/google/go-github/github"
	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// config prefix for github client setup
const configPrefix = "github"

const credentialsKey = "github-repo-pat"

var (
	errorVaultKeyNotExists = errors.New("unable to retrieve access token from vault: key does not exist")
	errorBadTokenFormat    = errors.New("unable to convert token from vault to string: bad format")
	errorUnknownAuthMethod = errors.New("unsupported github auth method, options are local | vault")
)

type Client struct {
	client *github.Client
}

// githubConfig contains configuration for initializing a github api client
type githubConfig struct {
	// AccessToken is used to reader a github Personal Access Token out of an environment variable,
	// otherwise one will be pulled from vault
	Auth struct {
		Type  string `default:"local"`
		Vault struct {
			Path string `default:"secret/suitable/github/broadbot/tokens/ci-automation"`
			Key  string `default:"token"`
		}
	}
}

func New(options ...func(*Client) error) (*Client, error) {
	log.Debug().Msg("initializing github client")
	gh := &Client{}
	for _, option := range options {
		if err := option(gh); err != nil {
			return nil, err
		}
	}

	return gh, nil
}

func WithClient(client *http.Client) func(*Client) error {
	return func(c *Client) error {
		c.client = github.NewClient(client)
		return nil
	}
}

func WithDefaults(config config.Config, creds credentials.Credentials, vaultClientFactory func() (*vault.Client, error)) func(*Client) error {
	return func(c *Client) error {
		var cfg githubConfig
		if err := config.Unmarshal(configPrefix, &cfg); err != nil {
			return err
		}

		var tokenProvider credentials.TokenProvider
		if cfg.Auth.Type == "local" {
			log.Debug().Msg("using local github pat token provider")
			tokenProvider = buildLocalGithubTokenProvider(creds)
		} else if cfg.Auth.Type == "vault" {
			log.Debug().Msg("using vault github pat token provider")
			vault, err := vaultClientFactory()
			if err != nil {
				return err
			}
			tokenProvider = buildVaultGithubTokenProvider(creds, vault, cfg.Auth.Vault.Path, cfg.Auth.Vault.Key)
		} else {
			return errorUnknownAuthMethod
		}

		token, err := tokenProvider.Get()
		if err != nil {
			return err
		}

		ctx := context.Background()
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})
		client := oauth2.NewClient(ctx, tokenSource)
		c.client = github.NewClient(client)
		return nil
	}
}

func buildLocalGithubTokenProvider(creds credentials.Credentials) credentials.TokenProvider {
	return creds.NewTokenProvider(credentialsKey, func(options *credentials.TokenOptions) {
		options.PromptEnabled = true

		options.PromptMessage = `
A Github Personal Access Token is required in order to interact with the Github API.
You can generate a new PAT at https://github.com/settings/tokens (select ONLY the read:org scope).
Enter Personal Access Token: `

	})
}

func buildVaultGithubTokenProvider(creds credentials.Credentials, vault *vault.Client, path, key string) credentials.TokenProvider {
	return creds.NewTokenProvider(credentialsKey, func(options *credentials.TokenOptions) {
		options.IssueFn = func() ([]byte, error) {
			accessTokenSecret, err := vault.Logical().Read(path)
			if err != nil {
				return nil, err
			}

			tokenI, exists := accessTokenSecret.Data[key]
			if !exists {
				return nil, errorVaultKeyNotExists
			}

			// cast interface{} to string
			token, isByteSlice := tokenI.(string)
			if !isByteSlice {
				return nil, errorBadTokenFormat
			}

			return []byte(token), nil
		}
	})
}

func (c *Client) GetCallingUser(ctx context.Context) (string, error) {
	// passing "" to this method will turn user info for the authenticated user
	caller, _, err := c.client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return caller.GetLogin(), nil
}

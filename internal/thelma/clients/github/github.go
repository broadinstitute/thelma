package github

import (
	"context"
	"errors"
	"net/http"
	"time"

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
)

type Client struct {
	*github.Client
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

func New(config config.Config, creds credentials.Credentials, vaultClientFactory func() (*vault.Client, error)) (*Client, error) {
	log.Debug().Msg("initializing github client")
	var cfg githubConfig
	if err := config.Unmarshal(configPrefix, &cfg); err != nil {
		return nil, err
	}
	log.Debug().Msgf("config: %v", cfg)

	var tokenProvider credentials.TokenProvider
	if cfg.Auth.Type == "local" {
		tokenProvider = buildLocalGithubTokenProvider(creds)
	} else if cfg.Auth.Type == "vault" {
		vault, err := vaultClientFactory()
		if err != nil {
			return nil, err
		}
		tokenProvider = buildVaultGithubTokenProvider(creds, vault, cfg.Auth.Vault.Path, cfg.Auth.Vault.Key)
	}

	token, err := tokenProvider.Get()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})
	client := oauth2.NewClient(ctx, tokenSource)
	return &Client{github.NewClient(client)}, nil
}

func NewWithTestClient(testClient *http.Client) (*Client, error) {
	return &Client{github.NewClient(testClient)}, nil
}

func (c *Client) GetCallingUser(ctx context.Context) (string, error) {
	// passing "" to this method will turn user info for the authenticated user
	caller, _, err := c.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return caller.GetLogin(), nil
}

func buildLocalGithubTokenProvider(creds credentials.Credentials) credentials.TokenProvider {
	return creds.NewTokenProvider(credentialsKey, func(options *credentials.TokenOptions) {
		options.PromptEnabled = true

		options.PromptMessage = `
A Github Personal Access Token is required in order to interact with the Github API.
You can generate a new PAT at https://github.com/settings/tokens (select ONLY the read:org scope).
Enter Personal Access Token: `

		options.ValidateFn = func(token []byte) error {
			tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(token)})
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			client := oauth2.NewClient(ctx, tokenSource)
			gitClient := github.NewClient(client)
			// check if token is valid by getting calling user info
			_, _, err := gitClient.Users.Get(ctx, "")
			return err
		}
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

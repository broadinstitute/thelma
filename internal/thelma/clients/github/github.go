package github

import (
	"context"
	"errors"

	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/google/go-github/github"
	vault "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// config prefix for github client setup
const configPrefix = "github"

var (
	errorNoGithubPat       = errors.New("unable to determine PAT to use authenticating to Github: AccessToken is unset or empty")
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
	AccessToken *string
	Vault       struct {
		Path string `default:"secret/suitable/github/broadbot/tokens/ci-automation"`
		Key  string `default:"token"`
	}
}

func New(config config.Config, vaultClientFactory func() (*vault.Client, error)) (*Client, error) {
	log.Debug().Msg("initializing github client")
	var cfg githubConfig
	if err := config.Unmarshal(configPrefix, &cfg); err != nil {
		return nil, err
	}

	if utils.UnsetOrEmpty(cfg.AccessToken) {
		// if not set in environment then fall back to trying to grab a PAT from vault
		log.Debug().Msg("github PAT not found in env, attempting to fetch from vault")
		token, err := fetchAccessTokenFromVault(cfg.Vault.Path, cfg.Vault.Key, vaultClientFactory)
		if err != nil {
			return nil, err
		}
		cfg.AccessToken = &token
	}

	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *cfg.AccessToken})
	client := oauth2.NewClient(ctx, tokenSource)
	return &Client{github.NewClient(client)}, nil
}

func fetchAccessTokenFromVault(path, key string, vaultClientFactory func() (*vault.Client, error)) (string, error) {
	vault, err := vaultClientFactory()
	if err != nil {
		return "", err
	}

	accessTokenSecret, err := vault.Logical().Read(path)
	if err != nil {
		return "", err
	}

	tokenI, exists := accessTokenSecret.Data[key]
	if !exists {
		return "", errorVaultKeyNotExists
	}

	// cast interface{} to string
	token, isString := tokenI.(string)
	if !isString {
		return "", errorBadTokenFormat
	}

	return token, nil
}

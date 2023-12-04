package gha

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	configPrefix     = "gha"
	tokenProviderKey = "gha-oidc-token"
	// The default that TokenOptions would set up if we didn't do anything
	defaultTokenEnvVar = "GHA_OIDC_TOKEN"
	// The old value that the Sherlock client accepted directly, for backwards compatibility
	backwardsCompatibilityTokenEnvVar = "SHERLOCK_GHA_OIDC_TOKEN"
)

type ghaConfig struct {
	Oidc struct {
		// ValidatingIssuer is the host for the OIDC provider configuration, which is expected to be at
		// <ValidatingIssuer>/.well-known/openid-configuration
		ValidatingIssuer string `default:"https://token.actions.githubusercontent.com"`
		// Behavior controls whether the GHA OIDC token mechanism should be used:
		// - "always": always use it and return any errors doing so
		// - "never": never use it -- tokens returned by it will be empty but no errors will be returned either
		// - "opportunistic": use it if it's available, but if it's not, don't return an error, simply an empty token
		Behavior string `default:"opportunistic" one-of:"always never opportunistic"`
	}
}

// NewGhaOidcProvider returns a credentials.TokenProvider that will return a GHA OIDC token if one is available,
// based on Thelma's configuration. Errors will only be returned from this function if the configuration can't be
// parsed or if the verifier can't be initialized.
func NewGhaOidcProvider(config config.Config, creds credentials.Credentials) (credentials.TokenProvider, error) {
	var cfg ghaConfig
	if err := config.Unmarshal(configPrefix, &cfg); err != nil {
		return nil, err
	}

	if err := initOidcVerifier(cfg.Oidc.ValidatingIssuer); err != nil {
		return nil, err
	}

	return &ghaOidcProviderWithBehavior{
		delegate: creds.GetTokenProvider(tokenProviderKey, func(options *credentials.TokenOptions) {
			options.ValidateFn = verifyOidcToken
			options.IssueFn = getOidcToken
			options.EnvVars = []string{defaultTokenEnvVar, backwardsCompatibilityTokenEnvVar}
		}),
		behavior: cfg.Oidc.Behavior,
	}, nil
}

type ghaOidcProviderWithBehavior struct {
	delegate credentials.TokenProvider
	behavior string
}

func (g *ghaOidcProviderWithBehavior) Get() ([]byte, error) {
	switch g.behavior {
	case "always":
		return g.delegate.Get()
	case "never":
		return nil, nil
	case "opportunistic":
		if token, err := g.delegate.Get(); err != nil {
			log.Trace().Err(err).Msgf("%T swallowed error from Get() due to opportunistic behavior", g)
			return nil, nil
		} else {
			return token, nil
		}
	default:
		return nil, errors.Errorf("unknown %T behavior: %s", g, g.behavior)
	}
}

func (g *ghaOidcProviderWithBehavior) Reissue() ([]byte, error) {
	switch g.behavior {
	case "always":
		return g.delegate.Reissue()
	case "never":
		return nil, nil
	case "opportunistic":
		if token, err := g.delegate.Reissue(); err != nil {
			log.Trace().Err(err).Msgf("%T swallowed error from Reissue() due to opportunistic behavior", g)
			return nil, nil
		} else {
			return token, nil
		}
	default:
		return nil, errors.Errorf("unknown %T behavior: %s", g, g.behavior)
	}
}

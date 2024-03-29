package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
)

func googleProvider(creds credentials.Credentials, cfg iapConfig, googleClient google.Clients) (credentials.TokenProvider, error) {
	issuer, err := googleClient.IdTokenGenerator(cfg.Audience)
	if err != nil {
		return nil, err
	}
	return creds.GetTokenProvider(tokenKey, func(options *credentials.TokenOptions) {
		options.EnvVars = []string{defaultTokenEnvVar, backwardsCompatibilityTokenEnvVar}
		options.IssueFn = issuer
		options.ValidateFn = idtokenValidator
	}), nil
}

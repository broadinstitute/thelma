package newiap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
)

func googleProvider(creds credentials.Credentials, cfg iapConfig, googleClient google.Clients) (credentials.TokenProvider, error) {
	issuer, err := googleClient.IdTokenGenerator(cfg.ClientID)
	if err != nil {
		return nil, err
	}
	return creds.GetTokenProvider(tokenKey, func(options *credentials.TokenOptions) {
		options.IssueFn = issuer
		options.ValidateFn = makeIdTokenValidator(cfg)
	}), nil
}

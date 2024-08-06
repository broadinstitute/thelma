package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
)

func googleProvider(creds credentials.Credentials, cfg iapConfig, googleClient google.Clients, project Project) (credentials.TokenProvider, error) {
	clientID, _, err := project.oauthCredentials(cfg)
	if err != nil {
		return nil, err
	}
	tokenKey, err := project.tokenKey()
	if err != nil {
		return nil, err
	}

	issuer, err := googleClient.IdTokenGenerator(clientID)
	if err != nil {
		return nil, err
	}
	return creds.GetTokenProvider(tokenKey, func(options *credentials.TokenOptions) {
		options.IssueFn = issuer
		options.ValidateFn = makeIdTokenValidator(clientID)
	}), nil
}

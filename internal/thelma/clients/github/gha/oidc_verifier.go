package gha

import (
	"context"
	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

var verifier *oidc.IDTokenVerifier

// initOidcVerifier will initialize the verifier if it hasn't been initialized yet.
func initOidcVerifier(validatingIssuer string) error {
	if verifier == nil {
		provider, err := oidc.NewProvider(context.Background(), validatingIssuer)
		if err != nil {
			time.Sleep(time.Second)
			provider, err = oidc.NewProvider(context.Background(), validatingIssuer)
			if err != nil {
				return err
			} else {
				log.Info().Msg("Thelma recovered from a transient error while initializing the GHA OIDC verifier")
			}
		}

		type extraConfigurationClaims struct {
			IdTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
		}
		var claims extraConfigurationClaims
		if err = provider.Claims(&claims); err != nil {
			return err
		}

		verifier = provider.Verifier(&oidc.Config{
			// The ClientID gets compared to the "aud" claim of the returned OIDC token. GitHub actually allows
			// customization of the "aud" claim, so we don't rely on it.
			SkipClientIDCheck: true,
			// The library says it defaults to RS256, but GitHub includes this information at its configuration
			// endpoint, so we'll grab it to be safe.
			SupportedSigningAlgs: claims.IdTokenSigningAlgValuesSupported,
		})
	}
	return nil
}

func verifyOidcToken(token []byte) error {
	if verifier == nil {
		return errors.Errorf("verifier was nil, was initOidcVerifier called?")
	} else {
		_, err := verifier.Verify(context.Background(), string(token))
		return err
	}
}

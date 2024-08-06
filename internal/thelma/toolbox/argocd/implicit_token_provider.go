package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/utils/serve_redirect"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"net/http"
)

const (
	tokenStorageKey = "argocd-token"
	redirectPort    = 8085             // Same as the CLI's localhost callback URL
	redirectPath    = "/auth/callback" // Same as the CLI's localhost callback URL
)

var scopes = []string{oidc.ScopeOpenID, "profile", "email", "groups"} // Can't store slice as const

func implicitTokenProvider(creds credentials.Credentials, cfg argocdConfig, sherlockHttpClient *http.Client) (credentials.TokenProvider, error) {
	ctx := context.Background()
	ctx = oidc.ClientContext(ctx, sherlockHttpClient) // Use the client for all requests made by oauth2/oidc packages
	oidcProvider, providerErr := oidc.NewProvider(ctx, cfg.SherlockOidcProvider)
	if providerErr != nil {
		return nil, errors.Errorf("failed to create oidc provider: %v", providerErr)
	}
	oauth2Config := oauth2.Config{
		ClientID:    cfg.SherlockOidcCliClientID, // No client secret, we're using PKCE
		RedirectURL: fmt.Sprintf("http://localhost:%d%s", redirectPort, redirectPath),
		Endpoint:    oidcProvider.Endpoint(),
		Scopes:      scopes,
	}
	oidcVerifier := oidcProvider.VerifierContext(ctx, &oidc.Config{ClientID: cfg.SherlockOidcCliClientID})

	type storedFormat struct {
		Token   *oauth2.Token `json:"token"`
		IdToken string        `json:"idToken"`
	}

	return credentials.GetTypedTokenProvider(creds, tokenStorageKey, func(options *credentials.TypedTokenOptions[*oauth2.Token]) {
		options.UnmarshalFromStoreFn = func(bytes []byte) (*oauth2.Token, error) {
			var stored storedFormat
			if err := json.Unmarshal(bytes, &stored); err != nil {
				return nil, err
			} else if stored.Token == nil {
				return nil, errors.Errorf("stored *oauth2.Token was nil (perhaps it came from an older version of thelma?)")
			}
			return stored.Token.WithExtra(map[string]interface{}{"id_token": stored.IdToken}), nil
		}
		options.MarshalToStoreFn = func(token *oauth2.Token) ([]byte, error) {
			return json.Marshal(&storedFormat{
				Token:   token,
				IdToken: token.Extra("id_token").(string),
			})
		}
		options.MarshalToReturnFn = func(token *oauth2.Token) ([]byte, error) {
			if idtoken, ok := token.Extra("id_token").(string); !ok {
				return nil, errors.Errorf("id token was unexpected type %T", token.Extra("id_token"))
			} else {
				return []byte(idtoken), nil
			}
		}
		options.RefreshFn = func(token *oauth2.Token) (*oauth2.Token, error) {
			return oauth2Config.TokenSource(ctx, token).Token()
		}
		options.ValidateFn = func(token *oauth2.Token) error {
			if idtoken, ok := token.Extra("id_token").(string); !ok {
				return errors.Errorf("id token was unexpected type %T", token.Extra("id_token"))
			} else {
				_, err := oidcVerifier.Verify(ctx, idtoken)
				return err
			}
		}
		options.IssueFn = func() (*oauth2.Token, error) {
			log.Debug().Msg("Generating ArgoCD auth based on Sherlock...")
			// Start server to receive the redirect
			var authorizationCode string
			state, closeFunc, err := serve_redirect.ServeRedirect(redirectPort, redirectPath, func(code string) {
				authorizationCode = code
			})
			if err != nil {
				return nil, errors.Errorf("failed to serve redirect: %v", err)
			}
			defer closeFunc()

			// Get the authorization code with PKCE: we'll get redirected from /oidc/authorize to /login to
			// /oidc/authorize/callback to our own redirect URL we're serving above. This is like normal
			// OAuth except the Sherlock's /login merely needs our auth in headers, so it doesn't actually
			// require any interaction. We just get redirected all the way through.
			oauth2Verifier := oauth2.GenerateVerifier()
			url := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(oauth2Verifier))
			response, err := sherlockHttpClient.Get(url)
			if err != nil {
				return nil, errors.Errorf("failed to get URL: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				return nil, errors.Errorf("unexpected status code: %d", response.StatusCode)
			}
			if authorizationCode == "" {
				return nil, errors.New("authorization code not received")
			}

			// Now with the code, do the exchange to get the token we want
			tok, err := oauth2Config.Exchange(ctx, authorizationCode, oauth2.VerifierOption(oauth2Verifier))
			if err != nil {
				return nil, errors.Errorf("failed to exchange authorization code for token: %v", err)
			}
			log.Debug().Msg("ArgoCD auth generated")
			return tok, nil
		}
	})
}

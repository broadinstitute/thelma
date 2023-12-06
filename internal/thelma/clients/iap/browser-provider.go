package iap

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// how long to wait before timing out compute engine metadata request
	computeEngineMetadataRequestTimeout = 15 * time.Second
)

func browserProvider(creds credentials.Credentials, cfg iapConfig, vaultClient *vaultapi.Client, runner shell.Runner) (credentials.TokenProvider, error) {
	oauthConfig, redirectPort, err := createOAuthConfig(cfg, vaultClient)
	if err != nil {
		return nil, err
	}

	type storedFormat struct {
		Token   *oauth2.Token `json:"token"`
		IdToken string        `json:"idToken"`
	}

	return credentials.GetTypedTokenProvider(creds, tokenKey, func(options *credentials.TypedTokenOptions[*oauth2.Token]) {
		options.EnvVars = []string{defaultTokenEnvVar, backwardsCompatibilityTokenEnvVar}
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
			return oauthConfig.TokenSource(context.Background(), token).Token()
		}
		options.ValidateFn = func(token *oauth2.Token) error {
			if idtoken, ok := token.Extra("id_token").(string); !ok {
				return errors.Errorf("id token was unexpected type %T", token.Extra("id_token"))
			} else {
				return idtokenValidator([]byte(idtoken))
			}
		}
		options.IssueFn = func() (*oauth2.Token, error) {
			if authorizationCode, err := useBrowserForAuthorizationCode(oauthConfig, runner, redirectPort); err != nil {
				return nil, fmt.Errorf("unable to obtain authorization code via browser: %v", err)
			} else if token, err := oauthConfig.Exchange(context.Background(), authorizationCode); err != nil {
				return nil, fmt.Errorf("unable to exchange authorization code for token: %v", err)
			} else {
				return token, err
			}
		}
	})
}

func createOAuthConfig(cfg iapConfig, vaultClient *vaultapi.Client) (*oauth2.Config, int, error) {
	var oauthCredentialsFields struct {
		AuthProviderX509CertUrl string   `json:"auth_provider_x509_cert_url" mapstructure:"auth_provider_x509_cert_url"`
		AuthUri                 string   `json:"auth_uri" mapstructure:"auth_uri"`
		ClientId                string   `json:"client_id" mapstructure:"client_id"`
		ClientSecret            string   `json:"client_secret" mapstructure:"client_secret"`
		ProjectId               string   `json:"project_id" mapstructure:"project_id"`
		RedirectUris            []string `json:"redirect_uris" mapstructure:"redirect_uris"`
		TokenUri                string   `json:"token_uri" mapstructure:"token_uri"`
	}

	if secret, err := vaultClient.Logical().Read(cfg.OAuthCredentials.VaultPath); err != nil {
		return nil, 0, errors.Errorf("error retrieving OAuth client credentials from Vault: %v", err)
	} else if secret == nil {
		return nil, 0, errors.Errorf("error retrieving OAuth client credentials from Vault: no secret at %s", cfg.OAuthCredentials.VaultPath)
	} else if encodedCreds, exists := secret.Data[cfg.OAuthCredentials.VaultKey]; !exists {
		return nil, 0, errors.Errorf("OAuth client credential secret at %s has unexpected format (missing key %s)", cfg.OAuthCredentials.VaultPath, cfg.OAuthCredentials.VaultKey)
	} else if _, isMap := encodedCreds.(map[string]interface{}); !isMap {
		return nil, 0, errors.Errorf("OAuth client credential secret at %s (key %s) has unexpected format (expected value to be map type)", cfg.OAuthCredentials.VaultPath, cfg.OAuthCredentials.VaultKey)
	} else if err = mapstructure.Decode(encodedCreds, &oauthCredentialsFields); err != nil {
		return nil, 0, errors.Errorf("error decoding OAuth client credentials: %v", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     oauthCredentialsFields.ClientId,
		ClientSecret: oauthCredentialsFields.ClientSecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	if redirectURI, err := selectAvailableLocalRedirectURI(oauthCredentialsFields.RedirectUris); err != nil {
		return nil, 0, err
	} else {
		oauthConfig.RedirectURL = redirectURI
	}

	parsedPortOfRedirectURI, err := parsePortOfRedirectURI(oauthConfig.RedirectURL)
	if err != nil {
		return nil, 0, err
	}

	return oauthConfig, parsedPortOfRedirectURI, nil
}

func selectAvailableLocalRedirectURI(redirectURIs []string) (string, error) {
	if len(redirectURIs) == 0 {
		return "", errors.Errorf("no redirect URIs provided")
	}

	var redirectURI string
	// Source of truth for redirect URI can't be in code/config, it must be in the OAuth Credentials
	// While we're at it might as well check that the URI port is open
	for _, uri := range redirectURIs {
		log.Trace().Msgf("checking redirect URI: %s", uri)
		parts := strings.Split(uri, ":")
		if len(parts) != 3 || parts[1] != "//localhost" {
			continue
		}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%s", parts[2]))
		_ = ln.Close()
		if err == nil {
			redirectURI = uri
			break
		} else {
			log.Trace().Msgf("couldn't use %s, (port busy)", uri)
		}
	}

	if redirectURI == "" {
		return "", errors.Errorf("unable to serve on any of the %d redirect URIs (ports busy?)", len(redirectURIs))
	}
	log.Trace().Msgf("Using redirect URI of %s", redirectURI)
	return redirectURI, nil
}

func parsePortOfRedirectURI(redirectURI string) (int, error) {
	if parsedRedirectURI, err := url.Parse(redirectURI); err != nil {
		return 0, errors.Errorf("unable to parse redirect URI to identify its port: %v", err)
	} else if stringPortOfRedirectURI := parsedRedirectURI.Port(); stringPortOfRedirectURI == "" {
		if parsedRedirectURI.Scheme == "http" {
			return 80, nil
		} else {
			return 443, nil
		}
	} else if portOfRedirectURI, err := strconv.Atoi(stringPortOfRedirectURI); err != nil {
		return 0, errors.Errorf("unable to parse redirect URI port: %v", err)
	} else {
		return portOfRedirectURI, nil
	}
}

func useBrowserForAuthorizationCode(config *oauth2.Config, runner shell.Runner, port int) (string, error) {
	log.Debug().Msgf("Obtaining OAuth authorization code...")
	stateBytes := make([]byte, 32)
	_, err := rand.Read(stateBytes)
	if err != nil {
		return "", err
	}
	// And that's when I realized I was CSRF-protecting a CLI
	state := base32.StdEncoding.EncodeToString(stateBytes)[:24]

	var authorizationCode string

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("code") == "" {
			// Browsers insist on looking for a /favicon.ico, so ignore it silently
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(writer, "%d - no code in request", http.StatusBadRequest)
		} else {
			log.Debug().Msgf("Received redirect with authorization code")
			if request.URL.Query().Get("state") != state {
				log.Debug().Msgf("Redirect state incorrect, rejecting (%s)", request.URL.String())
				writer.WriteHeader(http.StatusConflict)
				_, _ = fmt.Fprintf(writer, "%d - bad state", http.StatusConflict)
			} else {
				log.Debug().Msgf("Redirect state correct, storing")
				authorizationCode = request.URL.Query().Get("code")
				writer.WriteHeader(http.StatusOK)
				// Believe it or not, it is literally impossible for a page to close itself
				_, _ = fmt.Fprintf(writer, "Success! You can close this window.")
			}
		}
	})

	redirectServer := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	go func() {
		// We actually do want to precisely compare errors here, we're outputting at the trace level anyway
		//goland:noinspection GoDirectComparisonOfErrors
		if err := redirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Trace().Err(err).Msg("redirect server closed with unexpected error")
		}
	}()

	log.Trace().Msgf("redirect server available on port %d", port)
	browserUrl := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	var openBrowserCmd shell.Command
	switch runtime.GOOS {
	case "darwin":
		openBrowserCmd.Prog = "open"
		openBrowserCmd.Args = []string{browserUrl}
	case "windows":
		openBrowserCmd.Prog = "start"
		openBrowserCmd.Args = []string{browserUrl}
	default:
		// Has a ton of logic for how to handle Linux and other madness
		openBrowserCmd.Prog = "python3"
		openBrowserCmd.Args = []string{"-m", "webbrowser", browserUrl}
	}

	log.Debug().Msgf("using %s to launch browser on %s", openBrowserCmd.Prog, runtime.GOOS)
	log.Info().Msgf("Please visit the following URL in your web browser:\n\t%s", browserUrl)

	// Could blow up so we just let it fail silently; make the user copy-paste
	if err = runner.Run(openBrowserCmd); err != nil {
		log.Debug().Msgf("failed to open browser: %v", err)
	}
	for authorizationCode == "" {
		time.Sleep(50 * time.Millisecond)
	}
	log.Info().Msg("Received authorization code, thanks!")
	if err = redirectServer.Close(); err != nil {
		log.Trace().Msgf("error closing redirect server: %v", err)
	} else {
		log.Trace().Msgf("redirect server stopped")
	}
	return authorizationCode, nil
}

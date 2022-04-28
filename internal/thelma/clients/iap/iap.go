package iap

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//
// Authors:
// * Jack Warren: design & proof-of-concept
// * Chelsea Hoover: massaged into thelma
//

const configKey = "iap"
const credentialsKey = "iap-oauth-token"

type iapConfig struct {
	OAuthCredentialsVaultPath string `default:"secret/dsp/identity-aware-proxy/dsp-tools-k8s/dsp-tools-k8s-iap-oauth_client-credentials.json"`
	OAuthCredentialsVaultKey  string `default:"web"`
}

type oauthCredentials struct {
	AuthProviderX509CertUrl string   `json:"auth_provider_x509_cert_url" mapstructure:"auth_provider_x509_cert_url"`
	AuthUri                 string   `json:"auth_uri" mapstructure:"auth_uri"`
	ClientId                string   `json:"client_id" mapstructure:"client_id"`
	ClientSecret            string   `json:"client_secret" mapstructure:"client_secret"`
	ProjectId               string   `json:"project_id" mapstructure:"project_id"`
	RedirectUris            []string `json:"redirect_uris" mapstructure:"redirect_uris"`
	TokenUri                string   `json:"token_uri" mapstructure:"token_uri"`
}

// persistentToken is like oauth2.Token but with an IdToken field, which is what is used for IAP authentication
type persistentToken struct {
	AccessToken  string    `json:"access_token"`
	IdToken      string    `json:"id_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

// GetIDToken returns a valid IAP identity token, suitable for authenticating to DevOps services protected by IAP in the
// dsp-tools-k8s project
func GetIDToken(thelmaConfig config.Config, creds credentials.Credentials, vaultClient *vaultapi.Client) (string, error) {
	cfg, err := loadConfig(thelmaConfig)
	if err != nil {
		return "", err
	}

	oauthCreds, err := readOAuthClientCredentialsFromVault(vaultClient, cfg)
	if err != nil {
		return "", err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     oauthCreds.ClientId,
		ClientSecret: oauthCreds.ClientSecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	token := creds.NewToken(credentialsKey, func(options *credentials.TokenOptions) {
		options.IssueFn = func() ([]byte, error) {
			token, err := issueNewToken(oauthConfig, oauthCreds)
			if err != nil {
				return nil, err
			}
			return marshalPersistentToken(token)
		}
		options.ValidateFn = func(data []byte) error {
			token, err := unmarshalPersistentToken(data)
			if err != nil {
				return err
			}
			return validateToken(token, oauthConfig)
		}
	})

	content, err := token.Get()
	if err != nil {
		return "", err
	}
	oauthToken, err := unmarshalPersistentToken(content)
	if err != nil {
		return "", err
	}
	return oauthToken.Extra("id_token").(string), nil
}

func readOAuthClientCredentialsFromVault(vaultClient *vaultapi.Client, cfg iapConfig) (*oauthCredentials, error) {
	log.Debug().Msgf("Loading OAuth client credentials from %s", cfg.OAuthCredentialsVaultPath)
	secret, err := vaultClient.Logical().Read(cfg.OAuthCredentialsVaultPath)
	if err != nil {
		return nil, fmt.Errorf("error retrieving OAuth client credentials from Vault: %v", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("error retrieving OAuth client credentials from Vault: no secret at %s", cfg.OAuthCredentialsVaultPath)
	}

	encodedCreds, exists := secret.Data[cfg.OAuthCredentialsVaultKey]
	if !exists {
		return nil, fmt.Errorf("OAuth client credential secret at %s has unexpected format (missing key %s)", cfg.OAuthCredentialsVaultPath, cfg.OAuthCredentialsVaultKey)
	}
	_, isMap := encodedCreds.(map[string]interface{})
	if !isMap {
		return nil, fmt.Errorf("OAuth client credential secret at %s (key %s) has unexpected format (expected value to be map type)", cfg.OAuthCredentialsVaultPath, cfg.OAuthCredentialsVaultKey)
	}

	var oauthCreds oauthCredentials

	if err := mapstructure.Decode(encodedCreds, &oauthCreds); err != nil {
		return nil, fmt.Errorf("error decoding OAuth client credentials: %v", err)
	}

	return &oauthCreds, nil
}

// issue new token
func issueNewToken(oauthConfig *oauth2.Config, oauthCreds *oauthCredentials) (*oauth2.Token, error) {
	redirectURI, redirectPort, err := findRedirectURI(oauthCreds)
	if err != nil {
		return nil, err
	}
	oauthConfig.RedirectURL = redirectURI
	authorizationCode, err := obtainAuthorizationCode(redirectPort, oauthConfig)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("exchanging authorization code for access token...")
	token, err := oauthConfig.Exchange(context.Background(), authorizationCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// validate and possibly refresh token
func validateToken(token *oauth2.Token, oauthConfig *oauth2.Config) error {
	if idToken := token.Extra("id_token").(string); idToken == "" {
		return fmt.Errorf("token successfully read but lacked ID token")
	}
	tokenSource := oauthConfig.TokenSource(context.Background(), token)
	token, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("error validating token: %v", err)
	}
	if idToken := token.Extra("id_token").(string); idToken == "" {
		return fmt.Errorf("token lacked the ID token (unexpected behavior from oauth2 library or Google backend?)")
	}
	return nil
}

func unmarshalPersistentToken(data []byte) (*oauth2.Token, error) {
	var ptoken persistentToken
	err := json.Unmarshal(data, &ptoken)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling persistent token: %v", err)
	}

	token := &oauth2.Token{
		AccessToken:  ptoken.AccessToken,
		TokenType:    ptoken.TokenType,
		RefreshToken: ptoken.RefreshToken,
		Expiry:       ptoken.Expiry,
	}

	// propagate extra id_token field
	token = token.WithExtra(map[string]interface{}{"id_token": ptoken.IdToken})
	return token, nil
}

func marshalPersistentToken(token *oauth2.Token) ([]byte, error) {
	ptoken := &persistentToken{
		AccessToken:  token.AccessToken,
		IdToken:      token.Extra("id_token").(string),
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
	return json.Marshal(ptoken)
}

func loadConfig(thelmaConfig config.Config) (iapConfig, error) {
	var cfg iapConfig
	err := thelmaConfig.Unmarshal(configKey, &cfg)
	return cfg, err
}

func findRedirectURI(credentials *oauthCredentials) (string, int, error) {
	var redirectURI string
	// Source of truth for redirect URI can't be in code/config, it must be in the OAuth Credentials
	// While we're at it might as well check that the URI port is open
	for _, uri := range credentials.RedirectUris {
		log.Debug().Msgf("checking redirect URI: %s", uri)
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
			log.Debug().Msgf("couldn't use %s, (port busy)", uri)
		}
	}

	if redirectURI == "" {
		return "", 0, fmt.Errorf("unable to serve on any possible redirect URIs")
	}
	log.Debug().Msgf("Using redirect URI of %s", redirectURI)

	parsedRedirectURI, err := url.Parse(redirectURI)
	if err != nil {
		return "", 0, err
	}
	redirectPort, err := strconv.Atoi(parsedRedirectURI.Port())
	if err != nil {
		return "", 0, err
	}
	log.Debug().Msgf("will listen for redirects on port %d", redirectPort)
	return redirectURI, redirectPort, nil
}

func obtainAuthorizationCode(redirectPort int, oauthConfig *oauth2.Config) (string, error) {
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
				log.Debug().Msgf("Redirect state incorrect, rejecting", request.URL.String())
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

	redirectServer := &http.Server{Addr: fmt.Sprintf(":%d", redirectPort)}
	go func() {
		_ = redirectServer.ListenAndServe()
	}()

	log.Debug().Msgf("Redirect server available on port %d", redirectPort)
	browserUrl := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	log.Debug().Msgf("Using browser URL: %s", browserUrl)
	var openBrowserCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		openBrowserCmd = exec.Command("open", browserUrl)
	case "windows":
		openBrowserCmd = exec.Command("start", browserUrl)
	default:
		// Has a ton of logic for how to handle Linux and other madness
		openBrowserCmd = exec.Command("python3", "-m", "webbrowser", browserUrl)
	}

	log.Debug().Msgf("Using %s to launch browser on %s", openBrowserCmd.Path, runtime.GOOS)
	log.Info().Msgf("Please visit the following URL in your web browser:\n\t%s", browserUrl)
	// Could blow up so we just let it fail silently; make the user copy-paste
	if err := openBrowserCmd.Run(); err != nil {
		log.Debug().Msgf("Failed to open browser: %v", err)
	}
	for authorizationCode == "" {
		time.Sleep(200 * time.Millisecond)
	}
	log.Debug().Msgf("Authorization code stored")
	if err := redirectServer.Close(); err != nil {
		log.Debug().Msgf("Error closing redirect server: %v", err)
	}
	log.Debug().Msgf("Redirect server stopped")
	return authorizationCode, nil
}

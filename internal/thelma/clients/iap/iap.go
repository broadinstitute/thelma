package iap

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"net"
	"net/http"
	"net/url"
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

// configKey prefix used for configuration for this package
const configKey = "iap"

// tokenKey unique name for IAP tokens issued by this package, used to identify it in Thelma's token storage
const tokenKey = "iap-oauth-token"

// URL to request in order to validate IAP credentials are working
// Note that Sherlock doesn't actually have a thelma-iap-check endpoint so this will 404, but we don't care.
// We just care that we don't get the iap response header back in the response
const tokenValidationURL = "https://sherlock.dsp-devops.broadinstitute.org/thelma-iap-check"

// how long to wait before timing out token validation request
const tokenValidationRequestTimeout = 15 * time.Second

// Header returned by IAP indicating it intercepted the request and generated the response
const tokenValidationIapResponseHeader = "x-goog-iap-generated-response"

// how long to wait before timing out compute engine metadata request
const computeEngineMetadataRequestTimeout = 15 * time.Second

type iapConfig struct {
	Provider         string `default:"browser"  validate:"oneof=workloadidentity browser"`
	OAuthCredentials struct {
		VaultPath string `default:"secret/dsp/identity-aware-proxy/dsp-tools-k8s/dsp-tools-k8s-iap-oauth_client-credentials.json"`
		VaultKey  string `default:"web"`
	}
	WorkloadIdentity struct {
		ServiceAccount string `default:"default"` // default to using compute engine default service account
	}
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

// TokenProvider returns a new token provider for IAP tokens
func TokenProvider(thelmaConfig config.Config, creds credentials.Credentials, vaultClient *vaultapi.Client, runner shell.Runner) (credentials.TokenProvider, error) {
	cfg, err := loadConfig(thelmaConfig)
	if err != nil {
		return nil, err
	}

	oauthCreds, err := readOAuthClientCredentialsFromVault(vaultClient, cfg)
	if err != nil {
		return nil, err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     oauthCreds.ClientId,
		ClientSecret: oauthCreds.ClientSecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	// if workload identity is enabled, try to issue an IAP token that way first, falling back to user credentials
	if cfg.Provider == "workloadidentity" {
		return creds.NewTokenProvider(tokenKey, func(options *credentials.TokenOptions) {
			options.IssueFn = func() ([]byte, error) {
				return getTokenFromWorkloadIdentity(cfg, oauthConfig)
			}
			options.ValidateFn = func(token []byte) error {
				return validateIdentityToken(string(token))
			}
		}), nil
	}

	// else use browser provider
	if cfg.Provider == "browser" {
		provider := creds.NewTokenProvider(tokenKey, func(options *credentials.TokenOptions) {
			options.IssueFn = func() ([]byte, error) {
				token, err := issueNewToken(oauthConfig, oauthCreds, runner)
				if err != nil {
					return nil, err
				}
				return marshalPersistentToken(token)
			}
			options.RefreshFn = func(data []byte) ([]byte, error) {
				token, err := unmarshalPersistentToken(data)
				if err != nil {
					return nil, err
				}
				token, err = refreshToken(token, oauthConfig)
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
				return validateToken(token)
			}
		})

		return &tokenProvider{
			provider,
		}, nil
	}

	return nil, fmt.Errorf("unknown iap provider type: %v", cfg.Provider)
}

func readOAuthClientCredentialsFromVault(vaultClient *vaultapi.Client, cfg iapConfig) (*oauthCredentials, error) {
	log.Debug().Msgf("Loading OAuth client credentials from Vault (%s)", cfg.OAuthCredentials.VaultPath)
	secret, err := vaultClient.Logical().Read(cfg.OAuthCredentials.VaultPath)
	if err != nil {
		return nil, fmt.Errorf("error retrieving OAuth client credentials from Vault: %v", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("error retrieving OAuth client credentials from Vault: no secret at %s", cfg.OAuthCredentials.VaultPath)
	}

	encodedCreds, exists := secret.Data[cfg.OAuthCredentials.VaultKey]
	if !exists {
		return nil, fmt.Errorf("OAuth client credential secret at %s has unexpected format (missing key %s)", cfg.OAuthCredentials.VaultPath, cfg.OAuthCredentials.VaultKey)
	}
	_, isMap := encodedCreds.(map[string]interface{})
	if !isMap {
		return nil, fmt.Errorf("OAuth client credential secret at %s (key %s) has unexpected format (expected value to be map type)", cfg.OAuthCredentials.VaultPath, cfg.OAuthCredentials.VaultKey)
	}

	var oauthCreds oauthCredentials

	if err := mapstructure.Decode(encodedCreds, &oauthCreds); err != nil {
		return nil, fmt.Errorf("error decoding OAuth client credentials: %v", err)
	}

	return &oauthCreds, nil
}

func getTokenFromWorkloadIdentity(cfg iapConfig, oauthConfig *oauth2.Config) ([]byte, error) {
	metadataUrl := fmt.Sprintf("http://metadata/computeMetadata/v1/instance/service-accounts/%s/identity?audience=%s&format=full", cfg.WorkloadIdentity.ServiceAccount, oauthConfig.ClientID)
	log.Debug().Msgf("Attempting to issue new IAP token via workload identity")

	req, err := http.NewRequest(http.MethodGet, metadataUrl, nil)
	req.Header.Set("Metadata-Flavor", "Google")
	client := http.Client{
		Timeout: computeEngineMetadataRequestTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code from compute engine metadata: %v", resp.StatusCode)
	}
	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = resp.Body.Close(); err != nil {
		return nil, err
	}
	return token, nil
}

func issueNewToken(oauthConfig *oauth2.Config, oauthCreds *oauthCredentials, runner shell.Runner) (*oauth2.Token, error) {
	redirectURI, redirectPort, err := findRedirectURI(oauthCreds)
	if err != nil {
		return nil, err
	}
	oauthConfig.RedirectURL = redirectURI
	authorizationCode, err := obtainAuthorizationCode(redirectPort, oauthConfig, runner)
	if err != nil {
		return nil, err
	}

	log.Debug().Msgf("Exchanging authorization code for access token...")
	token, err := oauthConfig.Exchange(context.Background(), authorizationCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func refreshToken(token *oauth2.Token, oauthConfig *oauth2.Config) (*oauth2.Token, error) {
	tokenSource := oauthConfig.TokenSource(context.Background(), token)
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %v", err)
	}
	return token, nil
}

func validateToken(token *oauth2.Token) error {
	idToken := token.Extra("id_token").(string)
	if idToken == "" {
		return fmt.Errorf("token validation failed: id token is misssing")
	}
	return validateIdentityToken(idToken)
}

func validateIdentityToken(idToken string) error {
	// Build client
	client := http.Client{
		Timeout: tokenValidationRequestTimeout,
		// Don't follow IAP redirects
		// https://stackoverflow.com/questions/23297520/how-can-i-make-the-go-http-client-not-follow-redirects-automatically
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Build request
	req, err := http.NewRequest(http.MethodGet, tokenValidationURL, nil)
	if err != nil {
		return fmt.Errorf("error constructing validation request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", idToken))

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making validation request: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading token validation response body: %v", err)
	}
	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("error closing token validation response body: %v", err)
	}

	// Check for IAP header
	if resp.Header.Get(tokenValidationIapResponseHeader) != "" {
		return fmt.Errorf("token validation request was intercepted by IAP: %s (body: %q)", resp.Status, string(body))
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
	log.Debug().Msgf("Will listen for redirects on port %d", redirectPort)
	return redirectURI, redirectPort, nil
}

func obtainAuthorizationCode(redirectPort int, oauthConfig *oauth2.Config, runner shell.Runner) (string, error) {
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

	redirectServer := &http.Server{Addr: fmt.Sprintf(":%d", redirectPort)}
	go func() {
		_ = redirectServer.ListenAndServe()
	}()

	log.Debug().Msgf("Redirect server available on port %d", redirectPort)
	browserUrl := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	log.Debug().Msgf("Using browser URL: %s", browserUrl)
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

	log.Debug().Msgf("Using %s to launch browser on %s", openBrowserCmd.Prog, runtime.GOOS)
	log.Info().Msgf("Please visit the following URL in your web browser:\n\t%s", browserUrl)

	// Could blow up so we just let it fail silently; make the user copy-paste
	if err := runner.Run(openBrowserCmd); err != nil {
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

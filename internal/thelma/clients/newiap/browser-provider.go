package newiap

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"
)

const (
	// startingPort is the first port that will be tried when looking for an available port for the redirect URI.
	// It will count up from there.
	startingPort = 4444
	// portIterations is the number of ports that will be tried when looking for an available port for the redirect URI.
	// The redirect URI will be tried on ports starting from startingPort and counting up to startingPort + portIterations.
	portIterations = 10
)

func browserProvider(creds credentials.Credentials, cfg iapConfig, runner shell.Runner) (credentials.TokenProvider, error) {
	oauthConfig, redirectPort, err := createOAuthConfig(cfg)
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
				return nil, errors.Errorf("unable to obtain authorization code via browser: %v", err)
			} else if token, err := oauthConfig.Exchange(context.Background(), authorizationCode); err != nil {
				return nil, errors.Errorf("unable to exchange authorization code for token: %v", err)
			} else {
				return token, err
			}
		}
	})
}

func createOAuthConfig(cfg iapConfig) (*oauth2.Config, int, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	if redirectURI, err := selectAvailableLocalRedirectURI(); err != nil {
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

func selectAvailableLocalRedirectURI() (string, error) {
	for i := 0; i < portIterations; i++ {
		port := startingPort + i
		uri := fmt.Sprintf("http://localhost:%d", port)
		log.Trace().Msgf("checking redirect URI: %s", uri)
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if ln != nil {
			_ = ln.Close()
		}
		if err == nil {
			log.Trace().Msgf("using redirect URI: %s", uri)
			return uri, nil
		} else {
			log.Debug().Err(err).Msgf("couldn't use URI: %s", uri)
		}
	}
	return "", errors.New("unable to find an available local port for redirect URI")
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

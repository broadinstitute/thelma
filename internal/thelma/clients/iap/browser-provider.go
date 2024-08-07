package iap

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/utils/serve_redirect"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net"
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

func browserProvider(creds credentials.Credentials, cfg iapConfig, runner shell.Runner, project Project) (credentials.TokenProvider, error) {
	oauthConfig, redirectPort, err := createOAuthConfig(cfg, project)
	if err != nil {
		return nil, err
	}
	tokenKey, err := project.tokenKey()
	if err != nil {
		return nil, err
	}

	type storedFormat struct {
		Token   *oauth2.Token `json:"token"`
		IdToken string        `json:"idToken"`
	}

	idTokenValidator := makeIdTokenValidator(oauthConfig.ClientID)

	return credentials.GetTypedTokenProvider(creds, tokenKey, func(options *credentials.TypedTokenOptions[*oauth2.Token]) {
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
				return idTokenValidator([]byte(idtoken))
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

func createOAuthConfig(cfg iapConfig, project Project) (*oauth2.Config, int, error) {
	clientID, clientSecret, err := project.oauthCredentials(cfg)
	if err != nil {
		return nil, 0, err
	}
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
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
	var authorizationCode string
	state, closeFunc, err := serve_redirect.ServeRedirect(port, "/", func(code string) {
		authorizationCode = code
	})
	if err != nil {
		return "", errors.Errorf("failed to serve redirect: %v", err)
	}
	defer closeFunc()

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
	return authorizationCode, nil
}

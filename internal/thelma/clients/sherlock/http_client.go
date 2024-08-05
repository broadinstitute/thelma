package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

const sherlockGhaOidcHeader = "X-GHA-OIDC-JWT"

// HttpClient provides a reference to an authenticated http.Client that can be used to make requests to Sherlock.
// This is helpful mainly for things that the API Client isn't good at, like OIDC stuff.
func (c *clientImpl) HttpClient() *http.Client {
	return c.httpClient
}

// "sherlockRoundTripper implements http.RoundTripper"
var _ http.RoundTripper = &sherlockRoundTripper{}

type sherlockRoundTripper struct {
	// addrsToAuth avoids a potential vulnerability where Sherlock credentials could be sent to a non-Sherlock URL.
	// Sherlock would need to redirect us to a non-Sherlock URL for that to happen... but that's exactly what the
	// OIDC flow will try to do. Worth protecting against.
	addrsToAuth          []string
	iapTokenProvider     credentials.TokenProvider
	ghaOidcTokenProvider credentials.TokenProvider
	delegate             http.RoundTripper
}

func (s *sherlockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var shouldAuthorize bool
	for _, urlToAuthorize := range s.addrsToAuth {
		if strings.HasPrefix(req.URL.String(), urlToAuthorize) {
			shouldAuthorize = true
			break
		}
	}

	if shouldAuthorize {
		// IAP
		if s.iapTokenProvider != nil {
			if token, err := s.iapTokenProvider.Get(); err != nil {
				return nil, err
			} else if len(token) > 0 {
				req.Header.Set("Authorization", "Bearer "+string(token))
			}
		} else {
			return nil, errors.Errorf("IAP token provider is nil inside %T", s)
		}

		// GHA OIDC
		if s.ghaOidcTokenProvider != nil {
			if token, err := s.ghaOidcTokenProvider.Get(); err != nil {
				return nil, err
			} else if len(token) > 0 {
				req.Header.Set(sherlockGhaOidcHeader, string(token))
			}
		} else {
			return nil, errors.Errorf("GHA OIDC token provider is nil inside %T", s)
		}
	} else {
		// Don't log query parameters, needless security risk
		urlWithoutQuery, _, _ := strings.Cut(req.URL.String(), "?")
		log.Debug().Msgf("Sherlock HTTP client connecting to non-Sherlock host, omitting auth in request to %s", urlWithoutQuery)
	}

	return s.delegate.RoundTrip(req)
}

func makeHttpClient(
	addrsToAuth []string,
	iapTokenProvider credentials.TokenProvider,
	ghaOidcTokenProvider credentials.TokenProvider,
) *http.Client {
	return &http.Client{
		// Set this function so we can add logging because the redirects
		// are useful to know about when debugging OIDC
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Default behavior per docs
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			// Don't log query parameters, needless security risk
			urlWithoutQuery, _, _ := strings.Cut(req.URL.String(), "?")
			log.Debug().Msgf("Sherlock HTTP client following redirect to %s", urlWithoutQuery)
			return nil
		},
		Transport: &sherlockRoundTripper{
			addrsToAuth:          addrsToAuth,
			iapTokenProvider:     iapTokenProvider,
			ghaOidcTokenProvider: ghaOidcTokenProvider,
			delegate:             http.DefaultTransport,
		},
	}
}

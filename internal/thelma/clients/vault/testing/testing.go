// Package testing provides utilities for testing code that requires a Vault client
package testing

import (
	"encoding/json"
	"fmt"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const secretPrefix = "secret/"

// NewFakeVaultServer returns a new fake vault server that can be used to fake vault secret lookups
func NewFakeVaultServer(t *testing.T) *FakeVaultServer {
	_state := &state{
		secrets: make(map[string]map[string]interface{}),
	}

	mux := http.NewServeMux()

	mux.Handle("/v1/auth/github/login", toHttpHandler(_state.handleGithubLogin))
	mux.Handle("/v1/auth/token/lookup-self", toHttpHandler(_state.handleTokenLookup))
	mux.Handle("/v1/secret/", toHttpHandler(_state.handleSecret))
	mux.Handle("/", toHttpHandler(_state.handleUnmatchedRequest))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	return &FakeVaultServer{
		server: server,
		state:  _state,
	}
}

type FakeVaultServer struct {
	server *httptest.Server
	state  *state
}

// simplified http.Handler
type vaultApiHandler func(r *http.Request) (*vaultapi.Secret, error)

// represents state of the fake server
type state struct {
	secrets     map[string]map[string]interface{}
	expectLogin struct {
		enabled     bool
		githubToken string
		vaultToken  string
	}
}

// ConfigureClient can be used to configure a vault client to talk to this fake vault server instance
func (s *FakeVaultServer) ConfigureClient(clientConfig *vaultapi.Config) {
	clientConfig.HttpClient = s.server.Client()
	clientConfig.Address = s.server.URL
}

// GetClient can be used to quickly construct a vault client connected to this fake vault server instance
func (s *FakeVaultServer) GetClient() *vaultapi.Client {
	var cfg vaultapi.Config
	s.ConfigureClient(&cfg)
	client, err := vaultapi.NewClient(&cfg)
	if err != nil {
		panic(errors.Errorf("unexpected error configuring vault client for testing: %v", err))
	}
	return client
}

// Server returns the underlying httptest.Server associated with this fake vault server instance
func (s *FakeVaultServer) Server() *httptest.Server {
	return s.server
}

// ExpectGithubLogin configures the server to expect a github login with a specific Github token (by default any token is expected)
func (s *FakeVaultServer) ExpectGithubLogin(githubToken string, vaultToken string) {
	s.state.expectLogin.enabled = true
	s.state.expectLogin.githubToken = githubToken
	s.state.expectLogin.vaultToken = vaultToken
}

// SetSecret adds a secret to the fake server
func (s *FakeVaultServer) SetSecret(path string, data map[string]interface{}) {
	// remove secret/ prefix from key
	path = strings.TrimPrefix(path, secretPrefix)
	s.state.secrets[path] = data
}

func (s *state) handleGithubLogin(r *http.Request) (*vaultapi.Secret, error) {
	if r.Method != http.MethodPost &&
		r.Method != http.MethodPut {
		return nil, errors.Errorf("expected PUT or POST request")
	}

	var body struct {
		Token string `json:"token"`
	}

	if err := parseJsonRequestBody(r, &body); err != nil {
		return nil, err
	}

	if s.expectLogin.enabled {
		if body.Token != s.expectLogin.githubToken {
			return nil, errors.Errorf("github token mismatch: expected %q, got %q", s.expectLogin.githubToken, body.Token)
		}
	}

	return &vaultapi.Secret{
		Auth: &vaultapi.SecretAuth{
			ClientToken: s.expectLogin.vaultToken,
		},
	}, nil
}

func (s *state) handleTokenLookup(_ *http.Request) (*vaultapi.Secret, error) {
	return &vaultapi.Secret{
		Data: map[string]interface{}{
			"accessor": "00000000-0000-0000-0000-000000000000",
		},
	}, nil
}

func (s *state) handleSecret(r *http.Request) (*vaultapi.Secret, error) {
	secretPath := strings.TrimPrefix(r.URL.Path, "/v1/secret/")

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var data map[string]interface{}
		if err := parseJsonRequestBody(r, &data); err != nil {
			return nil, err
		}
		log.Debug().Msgf("setting secret %s to %v", secretPath, data)
		s.secrets[secretPath] = data

		var secret vaultapi.Secret
		secret.Data = data
		return &secret, nil
	}

	if r.Method == http.MethodGet {
		data, exists := s.secrets[secretPath]
		if !exists {
			log.Debug().Msgf("secret %s does not exist, returning 404", secretPath)
			return nil, nil
		}

		log.Debug().Msgf("returning secret %s: %v", secretPath, data)

		var secret vaultapi.Secret
		secret.Data = data
		return &secret, nil

	}

	return nil, errors.Errorf("invalid method for secrets api: %s", r.Method)
}

func (s *state) handleUnmatchedRequest(r *http.Request) (*vaultapi.Secret, error) {
	panic(errors.Errorf("no handler for request: %s %s", r.Method, r.URL.Path))
}

func writeSecretToResponseBody(secret *vaultapi.Secret, w http.ResponseWriter) {
	body, err := json.Marshal(secret)
	if err != nil {
		panic(errors.Errorf("error marshalling Vault secret to JSON (%v): %v", secret, err))
	}

	_, err = w.Write(body)
	if err != nil {
		panic(errors.Errorf("error writing response body: %v", err))
	}
}

// convert vaultApiHandler to http.Handler
func toHttpHandler(handler vaultApiHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msgf("%s %s", r.Method, r.URL.Path)

		secret, err := handler(r)

		if err != nil {
			log.Debug().Err(err).Msgf("400")

			http.Error(w, fmt.Sprintf("Bad request (%s %s): %v", r.Method, r.URL.Path, err), http.StatusBadRequest)
			return
		}

		if secret == nil {
			log.Debug().Msgf("404")

			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Debug().Interface("response", secret).Msgf("200")

		w.WriteHeader(http.StatusOK)
		writeSecretToResponseBody(secret, w)
	})
}

func parseJsonRequestBody(r *http.Request, into interface{}) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		panic(errors.Errorf("error reading request body: %v", err))
	}

	if err = json.Unmarshal(data, into); err != nil {
		return errors.Errorf("error unmarshalling request body: %v\n\n%s", err, string(data))
	}

	return nil
}

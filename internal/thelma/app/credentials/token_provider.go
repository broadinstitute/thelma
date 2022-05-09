package credentials

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/broadinstitute/thelma/internal/thelma/app/env"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
	"os"
	"strings"
)

// TokenOptions configuration options for a TokenProvider
type TokenOptions struct {
	// EnvVar (optional) environment variable to use for this token. Defaults to key (upper-cased with s/-/_/, eg. "vault-token" -> "VAULT_TOKEN")
	EnvVar string
	// PromptEnabled (optional) if true, user will be prompted to manually enter a token value if one does not exist in credential store.
	PromptEnabled bool
	// PromptMessage (optional) Override default prompt message ("Please enter VAULT_TOKEN: ")
	PromptMessage string
	// ValidateFn (optional) Optional function for validating a token. If supplied, stored credentials will be validated before being returned to caller
	ValidateFn func([]byte) error
	// RefreshFn (optional) Optional function for refreshing a token. Called if a stored credential turns out to be invalid. If an error is returned, IssueFn will be called to issue a new credential.
	RefreshFn func([]byte) ([]byte, error)
	// IssueFn (optional) Optional function for issuing a new token. If supplied, prompt options are ignored.
	IssueFn func() ([]byte, error)
	// CredentialStore (optional) Use a custom credential store instead of the default store (~/.thelma/credentials/$key)
	CredentialStore stores.Store
}

// TokenOption function for configuring a token's Options
type TokenOption func(*TokenOptions)

// TokenProvider manages a token used for authentication, possibly stored on the local filesystem
type TokenProvider interface {
	// Get returns the value of the token. Based on the token's options, it will attempt to resolve a value for
	// the token by:
	// (1) Looking it up in environment variables
	// (2) Looking it up in local credential store (~/.thelma/credentials)
	// (3) Issue a new token (if issuer function configured)
	// (4) Prompting user for value (if enabled)
	// If none of the token resolution options succeed an error is returned.
	Get() ([]byte, error)
	// Reissue forces re-issue of the token, without checking environment variables or for a valid existing
	// credential in the store
	Reissue() ([]byte, error)
}

// NewTokenProvider returns a new TokenProvider
func (c credentials) NewTokenProvider(key string, options ...TokenOption) TokenProvider {
	var opts TokenOptions
	for _, option := range options {
		option(&opts)
	}

	// set defaults if they were not set in option functions
	if opts.EnvVar == "" {
		opts.EnvVar = keyToEnvVar(key)
	}

	if opts.PromptMessage == "" {
		opts.PromptMessage = fmt.Sprintf("Please enter %s: ", opts.EnvVar)
	}

	if opts.CredentialStore == nil {
		opts.CredentialStore = c.defaultStore
	}

	return WithMasking(tokenProvider{
		key:     key,
		options: opts,
	})
}

type tokenProvider struct {
	key     string
	options TokenOptions
}

func (t tokenProvider) Get() ([]byte, error) {
	value := t.readFromEnv()
	if len(value) != 0 {
		return value, nil
	}

	value, err := t.readFromStore()
	if err != nil {
		return nil, err
	}
	if len(value) != 0 {
		return value, nil
	}

	return t.getNewToken()
}

func (t tokenProvider) Reissue() ([]byte, error) {
	return t.getNewToken()
}

// readFromEnv looks up a credential from the environment, checking with and without the THELMA_ prefix
// for example, ReadFromEnv("VAULT_TOKEN") will:
// (1) check for an environment variable THELMA_VAULT_TOKEN and return it if it exists
// (2) return the value of the VAULT_TOKEN environment variable
func (t tokenProvider) readFromEnv() []byte {
	value := os.Getenv(env.WithEnvPrefix(t.options.EnvVar))
	if len(value) != 0 {
		return []byte(value)
	}
	return []byte(os.Getenv(t.options.EnvVar))
}

// readFromStore looks up a token value in the credential store.
// If no value exists, the empty string is returned.
// If a value for the token exists it is not valid, readFromStore will attempt to refresh the token,
// returning the empty string if it can't be refreshed.
func (t tokenProvider) readFromStore() ([]byte, error) {
	exists, err := t.options.CredentialStore.Exists(t.key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	storedValue, err := t.options.CredentialStore.Read(t.key)
	if err != nil {
		return nil, err
	}

	err = t.validateToken(storedValue)
	if err != nil {
		log.Debug().Msgf("found value for %s in credential store, but validation function failed: %v", t.options.EnvVar, err)
		return t.refreshToken(storedValue)
	}

	return storedValue, nil
}

// refreshToken - return nil, nil if the token could not be refreshed or refresh failed
// returns an error if an exception (eg. error writing to credential store) occurs
// returns a non-nil value and no error if the token was successfully refreshed
func (t tokenProvider) refreshToken(value []byte) ([]byte, error) {
	if t.options.RefreshFn == nil {
		return nil, nil
	}

	log.Debug().Msgf("attempting to refresh token for %s", t.options.EnvVar)
	newValue, err := t.options.RefreshFn(value)
	if err != nil {
		log.Debug().Msgf("failed to refresh token %s: %v", t.options.EnvVar, err)
		return nil, nil
	}

	if err = t.validateToken(newValue); err != nil {
		// if this happens, there's likely a bug in the refresh function, so return an error
		return nil, fmt.Errorf("refresh for %s returned invalid token: %v", t.options.EnvVar, err)
	}

	log.Debug().Msgf("writing refreshed token %s to credential store", t.options.EnvVar)

	if err = t.options.CredentialStore.Write(t.key, newValue); err != nil {
		return nil, fmt.Errorf("error writing refreshed token %s to credential store: %v", t.options.EnvVar, err)
	}

	return newValue, nil
}

func (t tokenProvider) validateToken(value []byte) error {
	if t.options.ValidateFn == nil {
		// no validation function provided, assume value is valid
		return nil
	}

	return t.options.ValidateFn(value)
}

// getNewToken will attempt to get a new token value by either
// (1) invoking the issueFn callback
// (2) prompting the user for input
// If a new value is successfully obtained (and validated), token is stored and return to user
func (t tokenProvider) getNewToken() ([]byte, error) {
	var value []byte
	var err error

	if t.options.IssueFn != nil {
		log.Info().Msgf("Attempting to issue new %s", t.options.EnvVar)
		value, err = t.options.IssueFn()
	} else if t.options.PromptEnabled {
		value, err = t.promptForNewValue()
	} else {
		return nil, fmt.Errorf("could not issue new %s, no issueFn configured and input prompting is disabled", t.options.EnvVar)
	}

	if err != nil || len(value) == 0 {
		return value, err
	}

	err = t.validateToken(value)
	if err != nil {
		return nil, fmt.Errorf("new credential for %s is invalid: %v", t.options.EnvVar, err)
	}

	if err := t.options.CredentialStore.Write(t.key, value); err != nil {
		return nil, fmt.Errorf("failed to save new token value for %s: %v", t.options.EnvVar, err)
	}

	return value, nil
}

// promptForNewValue will prompt the user for a new token value
func (t tokenProvider) promptForNewValue() ([]byte, error) {
	if !utils.Interactive() {
		return nil, fmt.Errorf("can't prompt for %s (shell is not interactive), try passing in via environment variable %s", t.options.EnvVar, t.options.EnvVar)
	}

	fmt.Print(t.options.PromptMessage)
	value, err := term.ReadPassword(int(os.Stdin.Fd()))
	// print empty newline since ReadPassword doesn't
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("error reading user input for credential %s: %v", t.options.EnvVar, err)
	}

	return value, nil
}

func keyToEnvVar(key string) string {
	s := strings.ReplaceAll(key, "-", "_")
	return strings.ToUpper(s)
}

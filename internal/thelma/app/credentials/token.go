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

// TokenOptions configuration options for a Token
type TokenOptions struct {
	// EnvVar (optional) environment variable to use for this token. Defaults to key (upper-cased with s/-/_/, eg. "iap-token" -> "IAP_TOKEN")
	EnvVar string
	// PromptEnabled (optional) if true, user will be prompted to manually enter a token value if one does not exist in credential store.
	PromptEnabled bool
	// PromptMessage (optional) Override default prompt message ("Please enter IAP_TOKEN: ")
	PromptMessage string
	// ValidateFn (optional) Optional function for validating a token. If supplied, stored credentials will be validated before being returned to caller
	ValidateFn func([]byte) error
	// IssueFn (optional) Optional function for issuing a new token. If supplied, prompt options are ignored.
	IssueFn func() ([]byte, error)
	// CredentialStore (optional) Use a custom credential store instead of the default store (~/.thelma/credentials/$key)
	CredentialStore stores.Store
}

// TokenOption function for configuring a token's Options
type TokenOption func(*TokenOptions)

// Token represents a token used for authentication, possibly stored on the local filesystem
type Token interface {
	// Get returns the value of the token. Based on the token's options, it will attempt to resolve a value for
	// the token by:
	// (1) Looking it up in environment variables
	// (2) Looking it up in local credential store (~/.thelma/credentials)
	// (3) Issue a new token (if issuer function configured)
	// (4) Prompting user for value (if enabled)
	// If none of the token resolution options succeed an error is returned.
	Get() ([]byte, error)
}

// NewToken returns a new Token
func (c credentials) NewToken(key string, options ...TokenOption) Token {
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
		opts.CredentialStore = c.store
	}

	return token{
		key:     key,
		options: opts,
	}
}

type token struct {
	key     string
	options TokenOptions
}

func (t token) Get() ([]byte, error) {
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

// readFromEnv looks up a credential from the environment, checking with and without the THELMA_ prefix
// for example, ReadFromEnv("VAULT_TOKEN") will:
// (1) check for an environment variable THELMA_VAULT_TOKEN and return it if it exists
// (2) return the value of the VAULT_TOKEN environment variable
func (t token) readFromEnv() []byte {
	value := os.Getenv(env.WithEnvPrefix(t.options.EnvVar))
	if len(value) != 0 {
		return []byte(value)
	}
	return []byte(os.Getenv(t.options.EnvVar))
}

// readFromStore looks up a token value in the credential store. If no value exists, or the token value
// exists but the token's ValidateFn indicates the token value is no longer valid, the empty string is returned
func (t token) readFromStore() ([]byte, error) {
	exists, err := t.options.CredentialStore.Exists(t.key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, err
	}

	storedValue, err := t.options.CredentialStore.Read(t.key)
	if err != nil {
		return nil, err
	}

	if t.options.ValidateFn != nil {
		err := t.options.ValidateFn(storedValue)
		if err != nil {
			log.Debug().Msgf("found value for %s in credential store, but validation function failed: %v", t.options.EnvVar, err)
			return nil, nil
		}
	}

	return storedValue, nil
}

// getNewToken will attempt to get a new token value by either
// (1) invoking the issueFn callback
// (2) prompting the user for input
// If a new value is successfully obtained (and validated), the user will
func (t token) getNewToken() ([]byte, error) {
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

	if t.options.ValidateFn != nil {
		err := t.options.ValidateFn(value)
		if err != nil {
			return nil, fmt.Errorf("new credential for %s is invalid: %v", t.options.EnvVar, err)
		}
	}

	if err := t.options.CredentialStore.Write(t.key, value); err != nil {
		return nil, fmt.Errorf("failed to save new token value for %s: %v", t.options.EnvVar, err)
	}

	return value, nil
}

// promptForNewValue will prompt the user for a new token value
func (t token) promptForNewValue() ([]byte, error) {
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

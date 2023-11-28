package credentials

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/broadinstitute/thelma/internal/thelma/app/env"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
	"os"
	"strings"
	"sync"
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

	return withMasking(&tokenProvider{
		key:     key,
		options: opts,
	})
}

type tokenProvider struct {
	key     string
	options TokenOptions
	mutex   sync.RWMutex
}

func (t *tokenProvider) Get() ([]byte, error) {
	if value, err := t.getViaReadOnly(); err != nil {
		return nil, fmt.Errorf("%T.getViaReadOnly() error: %w", t, err)
	} else if len(value) > 0 {
		return value, nil
	} else if value, err = t.getViaReadWrite(); err != nil {
		return nil, fmt.Errorf("%T.getViaReadWrite() error: %w", t, err)
	} else {
		return value, nil
	}
}

func (t *tokenProvider) Reissue() ([]byte, error) {
	if err := t.resetViaReadWrite(); err != nil {
		return nil, fmt.Errorf("%T.resetViaReadWrite() error: %w", t, err)
	} else {
		return t.Get()
	}
}

// getViaReadOnly attempts to get a token, only reading. It may return nothing if a valid token wasn't readily available.
// It obtains a read lock on the tokenProvider and releases it before returning.
func (t *tokenProvider) getViaReadOnly() ([]byte, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if value, err := t.tryGetTokenOnlyReading(); err != nil {
		return nil, fmt.Errorf("%T.tryGetTokenOnlyReading() error: %w", t, err)
	} else if len(value) > 0 && t.validateToken(value) == nil {
		return value, nil
	} else {
		return nil, nil
	}
}

// getViaReadWrite gets a token, reading and possibly writing. It will always return either a token or an error.
// It obtains a read/write lock on the tokenProvider and releases it before returning.
func (t *tokenProvider) getViaReadWrite() ([]byte, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	// We read first in case another goroutine wrote while we were waiting for the lock,
	// also to get even a potentially invalid token to try to refresh.
	if value, err := t.tryGetTokenOnlyReading(); err != nil {
		return nil, fmt.Errorf("%T.tryGetTokenOnlyReading() error: %w", t, err)
	} else if len(value) > 0 {
		if err = t.validateToken(value); err == nil {
			return value, nil
		} else if value, err = t.tryGetAndWriteRefreshedToken(value); err != nil {
			return nil, fmt.Errorf("%T.tryGetAndWriteRefreshedToken() error: %w", t, err)
		} else if len(value) > 0 {
			return value, nil
		}
	}

	// If we get here, make a new token from scratch
	if value, err := t.mustGetAndWriteNewToken(); err != nil {
		return nil, fmt.Errorf("%T.mustGetAndWriteNewToken() error: %w", t, err)
	} else if len(value) > 0 {
		return value, nil
	} else {
		return nil, fmt.Errorf("%T.mustGetAndWriteNewToken() returned no error but no token either", t)
	}
}

// resetViaReadWrite resets internal state.
// It obtains a read/write lock on the tokenProvider and releases it before returning.
func (t *tokenProvider) resetViaReadWrite() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if exists, err := t.options.CredentialStore.Exists(t.key); err != nil {
		return fmt.Errorf("%T.Exists(%q) error: %w", t.options.CredentialStore, t.key, err)
	} else if exists {
		if err = t.options.CredentialStore.Remove(t.key); err != nil {
			return fmt.Errorf("%T.Remove(%q) error: %w", t.options.CredentialStore, t.key, err)
		}
	}
	return nil
}

// tryGetTokenOnlyReading attempts to get a token with only read access. It does not validate the token.
// It assumes the caller has locked the tokenProvider.
// It may return nothing if a token wasn't readily available.
func (t *tokenProvider) tryGetTokenOnlyReading() ([]byte, error) {
	// Short-circuit if we find a token in the environment
	for _, envVariableToCheck := range []string{env.WithEnvPrefix(t.options.EnvVar), t.options.EnvVar} {
		if value := os.Getenv(envVariableToCheck); len(value) > 0 {
			log.Trace().
				Str("variable", envVariableToCheck).
				Str("key", t.key).
				Type("type", t).
				Msgf("os.Getenv(%q) returned a value for %s, short-circuiting", envVariableToCheck, t.key)
			return []byte(value), nil
		}
	}

	if existsInStore, err := t.options.CredentialStore.Exists(t.key); err != nil {
		return nil, fmt.Errorf("%T.Exists(%q) error: %w", t.options.CredentialStore, t.key, err)
	} else if !existsInStore {
		return nil, nil
	} else if value, err := t.options.CredentialStore.Read(t.key); err != nil {
		return nil, fmt.Errorf("%T.Read(%q) error: %w", t.options.CredentialStore, t.key, err)
	} else {
		return value, nil
	}
}

// tryGetAndWriteRefreshedToken attempts to get a refreshed token using read and write access. It validates the token.
// It assumes the caller has locked the tokenProvider.
// It may return nothing if it wasn't able to refresh the token.
func (t *tokenProvider) tryGetAndWriteRefreshedToken(value []byte) ([]byte, error) {
	if t.options.RefreshFn == nil {
		return nil, nil
	} else if newValue, err := t.options.RefreshFn(value); err != nil {
		log.Trace().
			Err(err).
			Str("key", t.key).
			Type("type", t).
			Msgf("RefreshFn(%T.Read(%q)) error: %v", t.options.CredentialStore, t.key, err)
		return nil, nil
	} else if err = t.validateToken(newValue); err != nil {
		return nil, fmt.Errorf("%T.validateToken(RefreshFn(%T.Read(%q))) error: %w", t, t.options.CredentialStore, t.key, err)
	} else if err = t.options.CredentialStore.Write(t.key, newValue); err != nil {
		return nil, fmt.Errorf("%T.Write(%q, /* ... /*)) error: %w", t.options.CredentialStore, t.key, err)
	} else {
		return newValue, nil
	}
}

// mustGetAndWriteNewToken makes a new token using read and write access. It validates the token.
// It assumes the caller has locked the tokenProvider.
// It will always return either a token or an error.
func (t *tokenProvider) mustGetAndWriteNewToken() ([]byte, error) {
	var value []byte
	var err error

	if t.options.IssueFn != nil {
		if value, err = t.options.IssueFn(); err != nil {
			err = fmt.Errorf("%T.IssueFn() for %s error: %w", t, t.key, err)
		} else if len(value) == 0 {
			err = fmt.Errorf("%T.IssueFn() for %s returned no error but no token either", t, t.key)
		}
	} else if t.options.PromptEnabled {
		if value, err = t.promptForNewValue(); err != nil {
			err = fmt.Errorf("%T.promptForNewValue() for %s error: %w", t, t.key, err)
		} else if value == nil {
			err = fmt.Errorf("%T.promptForNewValue() for %s returned no error but no token either (user entered empty value?)", t, t.key)
		}
	} else {
		return nil, fmt.Errorf("could not issue new %s token; no IssueFn set and input prompting is disabled", t.key)
	}

	if err != nil {
		return nil, err
	} else if err = t.validateToken(value); err != nil {
		return nil, fmt.Errorf("%T.validateToken(/* ... /*)) for %s error: %w", t, t.key, err)
	} else if err = t.options.CredentialStore.Write(t.key, value); err != nil {
		return nil, fmt.Errorf("%T.Write(%q, /* ... /*)) error: %w", t, t.key, err)
	} else {
		return value, nil
	}
}

func (t *tokenProvider) validateToken(value []byte) error {
	if t.options.ValidateFn == nil {
		return nil
	} else {
		return t.options.ValidateFn(value)
	}
}

// promptForNewValue will prompt the user for a new token value
func (t *tokenProvider) promptForNewValue() ([]byte, error) {
	if !utils.Interactive() {
		return nil, errors.Errorf("can't prompt for %s (shell is not interactive), try passing in via environment variable %s", t.key, t.options.EnvVar)
	}

	fmt.Print(t.options.PromptMessage)
	value, err := term.ReadPassword(int(os.Stdin.Fd()))
	// print empty newline since ReadPassword doesn't
	fmt.Println()
	if err != nil {
		return nil, errors.Errorf("error reading user input for credential %s: %v", t.key, err)
	}

	return value, nil
}

func keyToEnvVar(key string) string {
	s := strings.ReplaceAll(key, "-", "_")
	return strings.ToUpper(s)
}

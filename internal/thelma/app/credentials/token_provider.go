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
	BaseTokenOptions
	// ValidateFn (optional) Optional function for validating a token. If supplied, stored credentials will be validated before being returned to caller.
	// This function can be called quite frequently in Goroutine scenarios, so offline validation is ideal.
	ValidateFn func([]byte) error
	// RefreshFn (optional) Optional function for refreshing a token. Called if a stored credential turns out to be invalid. If an error is returned, IssueFn will be called to issue a new credential.
	RefreshFn func([]byte) ([]byte, error)
	// IssueFn (optional) Optional function for issuing a new token. If supplied, prompt options are ignored.
	IssueFn func() ([]byte, error)

	// transformForReturn is an internal-use-only option used by GetTypedTokenProvider. This function is called when
	// a token is returned to the caller of TokenProvider.Get. It does not affect the stored token, nor is it used
	// when short-circuiting (e.g. when a token is found in the environment).
	transformForReturn func([]byte) ([]byte, error)
}

type BaseTokenOptions struct {
	// EnvVars (optional) environment variables to use for this token. Defaults to key (upper-cased with s/-/_/, eg. "vault-token" -> "VAULT_TOKEN").
	// Ideally only one environment variable should be used, but multiple are supported for backwards compatibility.
	// Environment variables will be checked both with the "THELMA_" prefix and without.
	EnvVars []string
	// PromptEnabled (optional) if true, user will be prompted to manually enter a token value if one does not exist in credential store.
	PromptEnabled bool
	// PromptMessage (optional) Override default prompt message ("Please enter VAULT_TOKEN: ")
	PromptMessage string
	// CredentialStore (optional) Use a custom credential store instead of the default store (~/.thelma/credentials/$key)
	CredentialStore stores.Store
}

// TokenOption function for configuring a token's Options
type TokenOption func(*TokenOptions)

// TokenProvider manages a token used for authentication, possibly stored on the local filesystem.
// The exported methods are Goroutine-safe.
type TokenProvider interface {
	// Get provides a token value based on the configuration of the TokenProvider. The overall flow is as follows:
	//
	// 1. If a match for any TokenOptions.EnvVars is found, immediately return that value
	// 2. If a match for the key is found in the TokenOptions.CredentialStore:
	//    - If it is valid per TokenOptions.ValidateFn, return it
	//    - If it is invalid but TokenOptions.RefreshFn is provided, attempt to refresh the token and validate, store, and return it
	//      (errors from TokenOptions.RefreshFn will cause the flow to continue to step 3)
	// 3. If TokenOptions.IssueFn is provided, issue a new token and validate, store, and return it
	// 4. If TokenOptions.IssueFn isn't provided but TokenOptions.PromptEnabled is true and the session is interactive,
	//    prompt the user for a new token value and validate, store, and return it
	Get() ([]byte, error)
	// Reissue clears the state of the TokenProvider and then calls Get (which will usually then issue a new token).
	Reissue() ([]byte, error)
}

var (
	TokenProviders      = map[string]TokenProvider{}
	tokenProvidersMutex = sync.RWMutex{}
)

// GetTokenProvider (see docs on Credentials)
func (c credentials) GetTokenProvider(key string, options ...TokenOption) TokenProvider {
	tokenProvidersMutex.RLock()
	if tp, ok := TokenProviders[key]; ok {
		tokenProvidersMutex.RUnlock()
		return tp
	}
	tokenProvidersMutex.RUnlock()
	tokenProvidersMutex.Lock()
	defer tokenProvidersMutex.Unlock()
	if tp, ok := TokenProviders[key]; ok {
		return tp
	} else {
		var opts TokenOptions
		for _, option := range options {
			option(&opts)
		}

		// set defaults if they were not set in option functions
		if len(opts.EnvVars) == 0 {
			opts.EnvVars = []string{keyToEnvVar(key)}
		}

		if opts.PromptMessage == "" {
			opts.PromptMessage = fmt.Sprintf("Please enter %s: ", key)
		}

		if opts.CredentialStore == nil {
			opts.CredentialStore = c.defaultStore
		}

		if opts.transformForReturn == nil {
			opts.transformForReturn = func(value []byte) ([]byte, error) { return value, nil }
		}

		TokenProviders[key] = withMasking(&tokenProvider{
			key:     key,
			options: opts,
		})

		return TokenProviders[key]
	}
}

type tokenProvider struct {
	key     string
	options TokenOptions
	mutex   sync.RWMutex
}

func (t *tokenProvider) Get() ([]byte, error) {
	if value, err := t.getViaReadOnly(); err != nil {
		return nil, errors.Errorf("%T.getViaReadOnly() error: %v", t, err)
	} else if len(value) > 0 {
		return value, nil
	} else if value, err = t.getViaReadWrite(); err != nil {
		return nil, errors.Errorf("%T.getViaReadWrite() error: %v", t, err)
	} else {
		return value, nil
	}
}

func (t *tokenProvider) Reissue() ([]byte, error) {
	if err := t.resetViaReadWrite(); err != nil {
		return nil, errors.Errorf("%T.resetViaReadWrite() error: %v", t, err)
	} else {
		return t.Get()
	}
}

// getViaReadOnly attempts to get a token, only reading. It may return nothing if a valid token wasn't readily available.
// It obtains a read lock on the tokenProvider and releases it before returning.
func (t *tokenProvider) getViaReadOnly() ([]byte, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if value, shortCircuited, err := t.tryGetTokenOnlyReading(); err != nil {
		return nil, errors.Errorf("%T.tryGetTokenOnlyReading() error: %v", t, err)
	} else if shortCircuited {
		// When short-circuiting, we don't call transformForReturn
		return value, nil
	} else if len(value) > 0 && t.validateToken(value) == nil {
		return t.options.transformForReturn(value)
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
	if value, shortCircuited, err := t.tryGetTokenOnlyReading(); err != nil {
		return nil, errors.Errorf("%T.tryGetTokenOnlyReading() error: %v", t, err)
	} else if shortCircuited {
		return value, nil
	} else if len(value) > 0 {
		if err = t.validateToken(value); err == nil {
			return t.options.transformForReturn(value)
		} else if value, err = t.tryGetAndWriteRefreshedToken(value); err != nil {
			return nil, errors.Errorf("%T.tryGetAndWriteRefreshedToken() error: %v", t, err)
		} else if len(value) > 0 {
			return t.options.transformForReturn(value)
		}
	}

	// If we get here, make a new token from scratch
	if value, err := t.mustGetAndWriteNewToken(); err != nil {
		return nil, errors.Errorf("%T.mustGetAndWriteNewToken() error: %v", t, err)
	} else if len(value) > 0 {
		return t.options.transformForReturn(value)
	} else {
		return nil, errors.Errorf("%T.mustGetAndWriteNewToken() returned no error but no token either", t)
	}
}

// resetViaReadWrite resets internal state.
// It obtains a read/write lock on the tokenProvider and releases it before returning.
func (t *tokenProvider) resetViaReadWrite() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if exists, err := t.options.CredentialStore.Exists(t.key); err != nil {
		return errors.Errorf("%T.Exists(%q) error: %v", t.options.CredentialStore, t.key, err)
	} else if exists {
		if err = t.options.CredentialStore.Remove(t.key); err != nil {
			return errors.Errorf("%T.Remove(%q) error: %v", t.options.CredentialStore, t.key, err)
		}
	}
	return nil
}

// tryGetTokenOnlyReading attempts to get a token with only read access. It does not validate the token.
// It will return a true boolean if short-circuiting occurred.
// It assumes the caller has locked the tokenProvider.
// It may return nothing if a token wasn't readily available.
func (t *tokenProvider) tryGetTokenOnlyReading() ([]byte, bool, error) {
	// Short-circuit if we find a token in the environment
	envVarsToCheck := t.options.EnvVars
	for _, envVarNeedingPrefix := range t.options.EnvVars {
		envVarsToCheck = append(envVarsToCheck, env.WithEnvPrefix(envVarNeedingPrefix))
	}
	for _, envVarToCheck := range envVarsToCheck {
		if value := os.Getenv(envVarToCheck); len(value) > 0 {
			log.Trace().
				Str("variable", envVarToCheck).
				Str("key", t.key).
				Type("type", t).
				Msgf("os.Getenv(%q) returned a value for %s, short-circuiting", envVarToCheck, t.key)
			return []byte(value), true, nil
		}
	}

	if existsInStore, err := t.options.CredentialStore.Exists(t.key); err != nil {
		return nil, false, errors.Errorf("%T.Exists(%q) error: %v", t.options.CredentialStore, t.key, err)
	} else if !existsInStore {
		return nil, false, nil
	} else if value, err := t.options.CredentialStore.Read(t.key); err != nil {
		return nil, false, errors.Errorf("%T.Read(%q) error: %v", t.options.CredentialStore, t.key, err)
	} else {
		return value, false, nil
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
		return nil, errors.Errorf("%T.validateToken(RefreshFn(%T.Read(%q))) error: %v", t, t.options.CredentialStore, t.key, err)
	} else if err = t.options.CredentialStore.Write(t.key, newValue); err != nil {
		return nil, errors.Errorf("%T.Write(%q, /* ... /*)) error: %v", t.options.CredentialStore, t.key, err)
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
			err = errors.Errorf("%T.IssueFn() for %s error: %v", t, t.key, err)
		} else if len(value) == 0 {
			err = errors.Errorf("%T.IssueFn() for %s returned no error but no token either", t, t.key)
		}
	} else if t.options.PromptEnabled {
		if value, err = t.promptForNewValue(); err != nil {
			err = errors.Errorf("%T.promptForNewValue() for %s error: %v", t, t.key, err)
		} else if value == nil {
			err = errors.Errorf("%T.promptForNewValue() for %s returned no error but no token either (user entered empty value?)", t, t.key)
		}
	} else {
		return nil, errors.Errorf("could not issue new %s token; no IssueFn set and input prompting is disabled", t.key)
	}

	if err != nil {
		return nil, err
	} else if err = t.validateToken(value); err != nil {
		return nil, errors.Errorf("%T.validateToken(/* ... /*)) for %s error: %v", t, t.key, err)
	} else if err = t.options.CredentialStore.Write(t.key, value); err != nil {
		return nil, errors.Errorf("%T.Write(%q, /* ... /*)) error: %v", t, t.key, err)
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
		// Safe to access opts.EnvVars[0] since we set a default in GetTokenProvider
		return nil, errors.Errorf("can't prompt for %s (shell is not interactive), try passing in via environment variable %s", t.key, t.options.EnvVars[0])
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

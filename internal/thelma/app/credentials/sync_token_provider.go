package credentials

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/env"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
	"os"
	"sync"
)

type syncTokenProvider struct {
	key     string
	options TokenOptions
	mutex   sync.RWMutex
}

func (t *syncTokenProvider) Get() ([]byte, error) {
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

func (t *syncTokenProvider) Reissue() ([]byte, error) {
	if err := t.resetViaReadWrite(); err != nil {
		return nil, fmt.Errorf("%T.resetViaReadWrite() error: %w", t, err)
	} else {
		return t.Get()
	}
}

// getViaReadOnly attempts to get a token, only reading. It may return nothing if a valid token wasn't readily available.
// It obtains a read lock on the syncTokenProvider and releases it before returning.
func (t *syncTokenProvider) getViaReadOnly() ([]byte, error) {
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
// It obtains a read/write lock on the syncTokenProvider and releases it before returning.
func (t *syncTokenProvider) getViaReadWrite() ([]byte, error) {
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
// It obtains a read/write lock on the syncTokenProvider and releases it before returning.
func (t *syncTokenProvider) resetViaReadWrite() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if err := t.options.CredentialStore.Remove(t.key); err != nil {
		return fmt.Errorf("%T.Remove(%q) error: %w", t.options.CredentialStore, t.key, err)
	} else {
		return nil
	}
}

// tryGetTokenOnlyReading attempts to get a token with only read access. It does not validate the token.
// It assumes the caller has locked the syncTokenProvider.
// It may return nothing if a token wasn't readily available.
func (t *syncTokenProvider) tryGetTokenOnlyReading() ([]byte, error) {
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
// It assumes the caller has locked the syncTokenProvider.
// It may return nothing if it wasn't able to refresh the token.
func (t *syncTokenProvider) tryGetAndWriteRefreshedToken(value []byte) ([]byte, error) {
	if t.options.RefreshFn == nil {
		return nil, nil
	} else if newValue, err := t.options.RefreshFn(value); err != nil {
		log.Trace().
			Err(err).
			Str("key", t.key).
			Type("type", t).
			Msgf("RefreshFn(%T.Read(%q)) error: %w", t.options.CredentialStore, t.key, err)
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
// It assumes the caller has locked the syncTokenProvider.
// It will always return either a token or an error.
func (t *syncTokenProvider) mustGetAndWriteNewToken() ([]byte, error) {
	var value []byte
	var err error

	if t.options.IssueFn != nil {
		if value, err = t.options.IssueFn(); err != nil {
			err = fmt.Errorf("%T.IssueFn() for %s error: %w", t, t.key, err)
		} else if value == nil {
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

func (t *syncTokenProvider) validateToken(value []byte) error {
	if t.options.ValidateFn == nil {
		return nil
	} else {
		return t.options.ValidateFn(value)
	}
}

// promptForNewValue will prompt the user for a new token value
func (t *syncTokenProvider) promptForNewValue() ([]byte, error) {
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

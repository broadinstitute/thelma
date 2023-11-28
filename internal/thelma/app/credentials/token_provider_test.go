package credentials

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/broadinstitute/thelma/internal/thelma/app/env"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"sync"
	"testing"
)

type mockStore struct {
	errorOnRead   bool
	bluffRead     string
	errorOnWrite  bool
	errorOnExists bool
	bluffExists   bool
	errorOnRemove bool
	delegate      stores.Store
}

func (s mockStore) Read(key string) ([]byte, error) {
	if s.errorOnRead {
		return nil, errors.Errorf("read error")
	} else if s.bluffRead != "" {
		return []byte(s.bluffRead), nil
	}
	return s.delegate.Read(key)
}

func (s mockStore) Exists(key string) (bool, error) {
	if s.errorOnExists {
		return false, errors.Errorf("exists error")
	} else if s.bluffExists {
		return true, nil
	}
	return s.delegate.Exists(key)
}

func (s mockStore) Write(key string, credential []byte) error {
	if s.errorOnWrite {
		return errors.Errorf("write error")
	}
	return s.delegate.Write(key, credential)
}

func (s mockStore) Remove(key string) error {
	if s.errorOnRemove {
		return errors.Errorf("remove error")
	}
	return s.delegate.Remove(key)
}

func Test_TokenProvider_Get(t *testing.T) {
	fakeEnvVar := fmt.Sprintf("FAKE_TOKEN_ENV_VAR_%d", os.Getpid())

	testCases := []struct {
		name        string
		key         string
		option      TokenOption
		setup       func(t *testing.T, tmpDir string)
		expectValue string
		expectErr   string
	}{
		{
			name:      "with defaults: should return error if token does not exist",
			key:       "my-token",
			expectErr: "^.*could not issue new my-token token; no IssueFn set and input prompting is disabled$",
		},
		{
			name: "with defaults: should return value in environment variable if defined",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.EnvVars = []string{fakeEnvVar}
			},
			setup: func(t *testing.T, tmpDir string) {
				err := os.Setenv(fakeEnvVar, "token-from-env")
				require.NoError(t, err)
			},
			expectValue: "token-from-env",
		},
		{
			name: "with defaults: should return value in environment variable if defined, picking from multiple",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.EnvVars = []string{fakeEnvVar, fakeEnvVar + "_2", fakeEnvVar + "_3"}
			},
			setup: func(t *testing.T, tmpDir string) {
				err := os.Setenv(fakeEnvVar, "token-from-env")
				require.NoError(t, err)
			},
			expectValue: "token-from-env",
		},
		{
			name: "with defaults: should return value in environment variable if defined with prefix",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.EnvVars = []string{fakeEnvVar}
			},
			setup: func(t *testing.T, tmpDir string) {
				err := os.Setenv(env.WithEnvPrefix(fakeEnvVar), "token-from-env")
				require.NoError(t, err)
			},
			expectValue: "token-from-env",
		},
		{
			name: "with defaults: should return token value if it exists",
			key:  "my-token",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("token-value"), 0600))
			},
			expectValue: "token-value",
		},
		{
			name: "with validateFn: should return error if token exists but is not valid",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.ValidateFn = func(v []byte) error {
					return errors.Errorf("this token is super invalid")
				}
			},
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("token-value"), 0600))
			},
			expectErr: "^.*could not issue new my-token token; no IssueFn set and input prompting is disabled$",
		},
		{
			name: "with issueFn: should issue new token if token does not exist",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.IssueFn = func() ([]byte, error) {
					return []byte("new-token-value"),
						nil
				}
			},
			expectValue: "new-token-value",
		},
		{
			name: "with issueFn: should return error if issueFn returns error",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.IssueFn = func() ([]byte, error) {
					return []byte{}, errors.Errorf("totally failed to issue new token")
				}
			},
			expectErr: "totally failed to issue new token",
		},
		{
			name: "with issueFn and validateFn: should issue new token if token exists but is not valid",
			key:  "my-token",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("old-token-value"), 0600))
			},
			option: func(options *TokenOptions) {
				options.IssueFn = func() ([]byte, error) {
					return []byte("new-token-value"), nil
				}
				options.ValidateFn = func(v []byte) error {
					if string(v) == "old-token-value" {
						return errors.Errorf("this token expired")
					}
					return nil
				}
			},
			expectValue: "new-token-value",
		},
		{
			name: "with issueFn and validateFn: should return error if new token is not valid",
			key:  "my-token",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("old-token-value"), 0600))
			},
			option: func(options *TokenOptions) {
				options.IssueFn = func() ([]byte, error) {
					return []byte("new-token-value"), nil
				}
				options.ValidateFn = func(_ []byte) error {
					return errors.Errorf("token is not valid for some reason")
				}
			},
			expectErr: "token is not valid for some reason",
		},
		{
			name: "with refreshFn and validateFn: should refresh token if token exists but is not valid",
			key:  "my-token",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("old-token-value"), 0600))
			},
			option: func(options *TokenOptions) {
				options.RefreshFn = func(_ []byte) ([]byte, error) {
					return []byte("refreshed-token-value"), nil
				}
				options.ValidateFn = func(v []byte) error {
					if string(v) == "old-token-value" {
						return errors.Errorf("this token expired")
					}
					return nil
				}
			},
			expectValue: "refreshed-token-value",
		},
		{
			name: "with refreshFn and validateFn: returns errors from writing newly refreshed token",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.RefreshFn = func(_ []byte) ([]byte, error) {
					return []byte("refreshed-token-value"), nil
				}
				options.ValidateFn = func(v []byte) error {
					if string(v) == "old-token-value" {
						return errors.Errorf("this token expired")
					}
					return nil
				}
				options.CredentialStore = mockStore{
					bluffExists:  true,
					bluffRead:    "old-token-value",
					errorOnWrite: true,
				}
			},
			expectErr: "^.*write error$",
		},
		{
			name: "with refreshFn and validateFn: should return error if refresh returns invalid token",
			key:  "my-token",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("old-token-value"), 0600))
			},
			option: func(options *TokenOptions) {
				options.RefreshFn = func(_ []byte) ([]byte, error) {
					return []byte("refreshed-token-value"), nil
				}
				options.ValidateFn = func(v []byte) error {
					return errors.Errorf("token is invalid")
				}
			},
			expectErr: "^.*token is invalid$",
		},
		{
			name: "with issueFn, refreshFn and validateFn: should issue new token if refresh fails",
			key:  "my-token",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("old-token-value"), 0600))
			},
			option: func(options *TokenOptions) {
				options.RefreshFn = func(_ []byte) ([]byte, error) {
					return nil, errors.Errorf("token too old to be refreshed")
				}
				options.IssueFn = func() ([]byte, error) {
					return []byte("new-token-value"), nil
				}
				options.ValidateFn = func(v []byte) error {
					if string(v) == "old-token-value" {
						return errors.Errorf("this token expired")
					}
					return nil
				}
			},
			expectValue: "new-token-value",
		},
		{
			name: "with issueFn, refreshFn and validateFn: returns errors from writing newly issued token",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.RefreshFn = func(_ []byte) ([]byte, error) {
					return nil, errors.Errorf("token too old to be refreshed")
				}
				options.IssueFn = func() ([]byte, error) {
					return []byte("new-token-value"), nil
				}
				options.ValidateFn = func(v []byte) error {
					if string(v) == "old-token-value" {
						return errors.Errorf("this token expired")
					}
					return nil
				}
				options.CredentialStore = mockStore{
					bluffExists:  true,
					bluffRead:    "old-token-value",
					errorOnWrite: true,
				}
			},
			expectErr: "^.*write error$",
		},
		{
			name: "with prompt enabled: return error because shell is not interactive",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.PromptEnabled = true
			},
			expectErr: "shell is not interactive",
		},
		{
			name: "with issueFn: errors if result is empty",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.IssueFn = func() ([]byte, error) {
					return []byte{}, nil
				}
			},
			expectErr: ".*returned no error but no token either.*",
		},
		{
			name: "returns errors from CredentialStore.Exists",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.CredentialStore = mockStore{
					errorOnExists: true,
					delegate:      stores.NewMapStore(),
				}
			},
			expectErr: "^.*exists error$",
		},
		{
			name: "returns errors from CredentialStore.Read",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.CredentialStore = mockStore{
					bluffExists: true,
					errorOnRead: true,
					delegate:    stores.NewMapStore(),
				}
			},
			expectErr: "^.*read error$",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeDir := t.TempDir()
			if tc.setup != nil {
				tc.setup(t, storeDir)
			}
			store, err := stores.NewDirectoryStore(storeDir)
			require.NoError(t, err)
			creds := NewWithStore(store)

			var options []TokenOption
			if tc.option != nil {
				options = append(options, tc.option)
			}
			tok := creds.NewTokenProvider(tc.key, options...)
			val, err := tok.Get()

			if tc.expectErr != "" {
				require.Error(t, err)
				assert.Regexp(t, tc.expectErr, err.Error())
				return
			}

			assert.Equal(t, tc.expectValue, string(val))
		})
	}
}

func Test_TokenProvider_Reissue(t *testing.T) {
	testCases := []struct {
		name        string
		key         string
		option      TokenOption
		setup       func(t *testing.T, tmpDir string)
		expectValue string
		expectErr   string
	}{
		{
			name: "with issueFn: should always issue new token even if valid token exists in credential store",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.IssueFn = func() ([]byte, error) {
					return []byte("new-token-value"), nil
				}
				options.ValidateFn = func(_ []byte) error {
					// both old and new tokens are valid
					return nil
				}
			},
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("old-token-value"), 0600))
			},
			expectValue: "new-token-value",
		},
		{
			name: "returns errors from CredentialStore.Exists",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.CredentialStore = mockStore{
					errorOnExists: true,
					delegate:      stores.NewMapStore(),
				}
			},
			expectErr: "^.*exists error$",
		},
		{
			name: "returns errors from CredentialStore.Remove",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.CredentialStore = mockStore{
					bluffExists:   true,
					errorOnRemove: true,
					delegate:      stores.NewMapStore(),
				}
			},
			expectErr: "^.*remove error$",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeDir := t.TempDir()
			if tc.setup != nil {
				tc.setup(t, storeDir)
			}
			store, err := stores.NewDirectoryStore(storeDir)
			require.NoError(t, err)
			creds := NewWithStore(store)

			var options []TokenOption
			if tc.option != nil {
				options = append(options, tc.option)
			}
			tok := creds.NewTokenProvider(tc.key, options...)
			val, err := tok.Reissue()

			if tc.expectErr != "" {
				require.Error(t, err)
				assert.Regexp(t, tc.expectErr, err.Error())
				return
			}

			assert.Equal(t, tc.expectValue, string(val))
		})
	}
}

// Test_TokenProvider_concurrency isn't perfect, but it would likely fail if TokenProvider weren't
// properly locking reads and writes. We're making an issuer that will fail if called more than once
// and validating that even with 100 concurrent goroutines, it only gets called once.
func Test_TokenProvider_concurrency(t *testing.T) {
	var issuerMutex sync.Mutex
	var issuerCalled bool
	issuerThatWillFailIfRunMoreThanOnce := func() ([]byte, error) {
		issuerMutex.Lock()
		defer issuerMutex.Unlock()
		if issuerCalled {
			return nil, fmt.Errorf("issuer already called")
		} else {
			issuerCalled = true
			return []byte("new-token-value"), nil
		}
	}

	storeDir := t.TempDir()
	store, err := stores.NewDirectoryStore(storeDir)
	require.NoError(t, err)
	creds := NewWithStore(store)
	tok := creds.NewTokenProvider("my-token", func(options *TokenOptions) {
		options.IssueFn = issuerThatWillFailIfRunMoreThanOnce
	})

	var wg sync.WaitGroup
	goroutineFn := func() {
		defer wg.Done()
		token, err := tok.Get()
		require.NoError(t, err)
		assert.Equal(t, "new-token-value", string(token))
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go goroutineFn()
	}
	wg.Wait()
}

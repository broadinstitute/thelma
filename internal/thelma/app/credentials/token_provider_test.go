package credentials

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_Token_Get(t *testing.T) {
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
			expectErr: "could not issue new MY_TOKEN",
		},
		{
			name: "with defaults: should return value in environment variable if defined",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.EnvVar = fakeEnvVar
			},
			setup: func(t *testing.T, tmpDir string) {
				err := os.Setenv(fakeEnvVar, "token-from-env")
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
					return fmt.Errorf("this token is super invalid")
				}
			},
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("token-value"), 0600))
			},
			expectErr: "could not issue new MY_TOKEN",
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
					return []byte{}, fmt.Errorf("totally failed to issue new token")
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
						return fmt.Errorf("this token expired")
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
					return fmt.Errorf("token is not valid for some reason")
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
						return fmt.Errorf("this token expired")
					}
					return nil
				}
			},
			expectValue: "refreshed-token-value",
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
					return fmt.Errorf("token is invalid")
				}
			},
			expectErr: "refresh for MY_TOKEN returned invalid token: token is invalid",
		},
		{
			name: "with issueFn, refreshFn and validateFn: should issue new token if refresh fails",
			key:  "my-token",
			setup: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.WriteFile(path.Join(tmpDir, "my-token"), []byte("old-token-value"), 0600))
			},
			option: func(options *TokenOptions) {
				options.RefreshFn = func(_ []byte) ([]byte, error) {
					return nil, fmt.Errorf("token too old to be refreshed")
				}
				options.IssueFn = func() ([]byte, error) {
					return []byte("new-token-value"), nil
				}
				options.ValidateFn = func(v []byte) error {
					if string(v) == "old-token-value" {
						return fmt.Errorf("this token expired")
					}
					return nil
				}
			},
			expectValue: "new-token-value",
		},
		{
			name: "with prompt enabled: return error because shell is not interactive",
			key:  "my-token",
			option: func(options *TokenOptions) {
				options.PromptEnabled = true
			},
			expectErr: "shell is not interactive",
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

func Test_Token_Reissue(t *testing.T) {
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

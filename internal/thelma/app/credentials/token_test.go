package credentials

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_Token_Get(t *testing.T) {
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
			creds, err := New(storeDir)
			require.NoError(t, err)

			var options []TokenOption
			if tc.option != nil {
				options = append(options, tc.option)
			}
			tok := creds.NewToken(tc.key, options...)
			val, err := tok.Get()

			if tc.expectErr != "" {
				require.Error(t, err)
				assert.Regexp(t, tc.expectErr, err.Error())
				return
			}

			assert.Equal(t, []byte(tc.expectValue), val)
		})
	}
}

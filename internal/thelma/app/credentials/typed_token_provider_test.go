package credentials

import (
	"encoding/json"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_TypedTokenProvider(t *testing.T) {
	type testToken struct {
		SavedField    string `json:"savedField"`
		ReturnedField string `json:"returnedField"`
	}

	var unmarshalShouldError, marshalShouldError, returnShouldError bool
	unmarshalFunc := func(fromStore []byte) (*testToken, error) {
		if unmarshalShouldError {
			return nil, errors.New("some error")
		}
		var tt testToken
		return &tt, json.Unmarshal(fromStore, &tt)
	}
	marshalFunc := func(tt *testToken) ([]byte, error) {
		if marshalShouldError {
			return nil, errors.New("some error")
		}
		return json.Marshal(tt)
	}
	returnFunc := func(tt *testToken) ([]byte, error) {
		if returnShouldError {
			return nil, errors.New("some error")
		}
		return []byte(tt.ReturnedField), nil
	}

	testCases := []struct {
		name                 string
		option               TypedTokenOption[*testToken]
		setup                func(t *testing.T, tmpDir string)
		unmarshalShouldError bool
		marshalShouldError   bool
		returnShouldError    bool
		expectCreationError  string
		expectValue          string
		expectErr            string
	}{
		{
			name: "unmarshal missing",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.MarshalToStoreFn = marshalFunc
			},
			expectCreationError: "TypedTokenOptions.UnmarshalFromStoreFn is required",
		},
		{
			name: "marshal missing",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
			},
			expectCreationError: "TypedTokenOptions.MarshalToStoreFn is required",
		},
		{
			name: "normal issue",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.IssueFn = func() (*testToken, error) {
					return &testToken{
						SavedField:    "some-saved-value",
						ReturnedField: "some-value",
					}, nil
				}
			},
			expectValue: "some-value",
		},
		{
			name: "normal issue with no special return fn",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.IssueFn = func() (*testToken, error) {
					return &testToken{
						SavedField:    "some-saved-value",
						ReturnedField: "some-value",
					}, nil
				}
			},
			expectValue: "{\"savedField\":\"some-saved-value\",\"returnedField\":\"some-value\"}",
		},
		{
			name: "issue that errors",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.IssueFn = func() (*testToken, error) {
					return nil, errors.New("some error")
				}
			},
			expectErr: ".*some error",
		},
		{
			name: "issue that can't marshal",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.IssueFn = func() (*testToken, error) {
					return &testToken{
						SavedField:    "some-saved-value",
						ReturnedField: "some-value",
					}, nil
				}
			},
			marshalShouldError: true,
			expectErr:          ".*couldn't marshal.*",
		},
		{
			name: "transform fails to unmarshal",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.CredentialStore = mockStore{
					bluffExists: true,
					bluffRead:   "{\"savedField\":\"some-saved-value\",\"returnedField\":\"some-value\"}",
				}
			},
			unmarshalShouldError: true,
			expectErr:            "error transforming token for return: couldn't parse.*",
		},
		{
			name: "transform fails to marshal",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.CredentialStore = mockStore{
					bluffExists: true,
					bluffRead:   "{\"savedField\":\"some-saved-value\",\"returnedField\":\"some-value\"}",
				}
			},
			returnShouldError: true,
			expectErr:         "error transforming token for return: couldn't marshal.*",
		},
		{
			name: "validate fails to unmarshal",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.IssueFn = func() (*testToken, error) {
					return &testToken{
						SavedField:    "some-saved-value",
						ReturnedField: "some-value",
					}, nil
				}
				o.ValidateFn = func(tt *testToken) error {
					return nil
				}
			},
			unmarshalShouldError: true,
			expectErr:            "error validating token: couldn't parse.*",
		},
		{
			name: "validate returns error",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.IssueFn = func() (*testToken, error) {
					return &testToken{
						SavedField:    "some-saved-value",
						ReturnedField: "some-value",
					}, nil
				}
				o.ValidateFn = func(tt *testToken) error {
					return errors.Errorf("some error")
				}
			},
			expectErr: ".*some error",
		},
		{
			name: "validate returns normally",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.CredentialStore = mockStore{
					bluffExists: true,
					bluffRead:   "{\"savedField\":\"some-saved-value\",\"returnedField\":\"some-value\"}",
				}
				o.ValidateFn = func(tt *testToken) error {
					if tt.SavedField != "some-saved-value" {
						return errors.Errorf("some error")
					} else {
						return nil
					}
				}
			},
			expectValue: "some-value",
		},
		{
			name: "refresh errors",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.CredentialStore = mockStore{
					bluffExists: true,
					bluffRead:   "{\"savedField\":\"some-old-value\",\"returnedField\":\"some-value\"}",
				}
				o.ValidateFn = func(tt *testToken) error {
					if tt.SavedField == "some-old-value" {
						return errors.Errorf("expired")
					} else {
						return nil
					}
				}
				o.RefreshFn = func(tt *testToken) (*testToken, error) {
					return nil, errors.Errorf("some error")
				}
			},
			expectErr: ".*no IssueFn set.*",
		},
		{
			name: "refresh returns normally",
			option: func(o *TypedTokenOptions[*testToken]) {
				o.UnmarshalFromStoreFn = unmarshalFunc
				o.MarshalToStoreFn = marshalFunc
				o.MarshalToReturnFn = returnFunc
				o.CredentialStore = mockStore{
					bluffExists: true,
					bluffRead:   "{\"savedField\":\"some-old-value\",\"returnedField\":\"some-value\"}",
					bluffWrite:  true,
				}
				o.ValidateFn = func(tt *testToken) error {
					if tt.SavedField == "some-old-value" {
						return errors.Errorf("expired")
					} else {
						return nil
					}
				}
				o.RefreshFn = func(tt *testToken) (*testToken, error) {
					return &testToken{
						SavedField:    "some-new-value",
						ReturnedField: "some-new-value",
					}, nil
				}
			},
			expectValue: "some-new-value",
		},
		// Can't test marshalling/unmarshalling inside refresh fn very well because those happen before when reading and
		// validating the token from the store and that would've been done already
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			unmarshalShouldError = tc.unmarshalShouldError
			marshalShouldError = tc.marshalShouldError
			returnShouldError = tc.returnShouldError
			storeDir := t.TempDir()
			if tc.setup != nil {
				tc.setup(t, storeDir)
			}
			store, err := stores.NewDirectoryStore(storeDir)
			require.NoError(t, err)
			creds := NewWithStore(store)

			var options []TypedTokenOption[*testToken]
			if tc.option != nil {
				options = append(options, tc.option)
			}
			tp, creationError := GetTypedTokenProvider[*testToken](creds, "my-token", options...)
			if tc.expectCreationError != "" {
				require.Error(t, creationError)
				assert.Regexp(t, tc.expectCreationError, creationError.Error())
				return
			}

			value, err := tp.Get()
			tokenProvidersMutex.RLock()
			TokenProviders = map[string]TokenProvider{}
			tokenProvidersMutex.RUnlock()

			if tc.expectErr != "" {
				require.Error(t, err)
				assert.Regexp(t, tc.expectErr, err.Error())
				return
			}

			assert.Equal(t, tc.expectValue, string(value))
		})
	}
}

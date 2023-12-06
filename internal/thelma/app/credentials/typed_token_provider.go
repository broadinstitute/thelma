package credentials

import "github.com/pkg/errors"

type TypedTokenOptions[T any] struct {
	BaseTokenOptions

	// UnmarshalFromStoreFn (required) parses T from the BaseTokenOptions.CredentialStore.
	//
	// If you're trying to change the stored representation of a token, it's safe to just error here.
	// UnmarshalFromStoreFn and MarshalToStoreFn get baked into TokenOptions.ValidateFn and similar,
	// so if they return an error it'll be seen as a validation error (and so will fall through to
	// refreshing/issuing). In other words, you don't need to gracefully handle it.
	UnmarshalFromStoreFn func([]byte) (T, error)
	// MarshalToStoreFn (required) will be used to store T in the BaseTokenOptions.CredentialStore.
	MarshalToStoreFn func(T) ([]byte, error)

	// MarshalToReturnFn (optional) will be used to return T to the caller of TokenProvider.Get.
	// If omitted, MarshalToStoreFn will be used.
	MarshalToReturnFn func(T) ([]byte, error)

	// ValidateFn (optional) is like TokenOptions.ValidateFn, but for T.
	ValidateFn func(T) error
	// RefreshFn (optional) is like TokenOptions.RefreshFn, but for T.
	RefreshFn func(T) (T, error)
	// IssueFn (optional) is like TokenOptions.IssueFn, but for T.
	IssueFn func() (T, error)
}

type TypedTokenOption[T any] func(*TypedTokenOptions[T])

// GetTypedTokenProvider is like Credentials.GetTokenProvider but for TypedTokenOptions.
// Go's lackluster generics support prevents this method from actually being present on
// Credentials, so the receiver must be passed here.
func GetTypedTokenProvider[T any](c Credentials, key string, options ...TypedTokenOption[T]) (TokenProvider, error) {
	// Use given options
	var typedOptions TypedTokenOptions[T]
	for _, typedOption := range options {
		typedOption(&typedOptions)
	}

	// Validate given options
	if typedOptions.UnmarshalFromStoreFn == nil {
		return nil, errors.Errorf("TypedTokenOptions.UnmarshalFromStoreFn is required")
	} else if typedOptions.MarshalToStoreFn == nil {
		return nil, errors.Errorf("TypedTokenOptions.MarshalToStoreFn is required")
	}

	// Set defaults
	if typedOptions.MarshalToReturnFn == nil {
		typedOptions.MarshalToReturnFn = typedOptions.MarshalToStoreFn
	}

	return c.GetTokenProvider(key, func(plainOptions *TokenOptions) {
		// Translate typed options to plain options
		if typedOptions.ValidateFn != nil {
			plainOptions.ValidateFn = func(fromStore []byte) error {
				t, err := typedOptions.UnmarshalFromStoreFn(fromStore)
				if err != nil {
					return errors.Errorf("error validating token: couldn't parse from []byte: %v", err)
				}

				return typedOptions.ValidateFn(t)
			}
		}
		if typedOptions.RefreshFn != nil {
			plainOptions.RefreshFn = func(fromStore []byte) ([]byte, error) {
				t, err := typedOptions.UnmarshalFromStoreFn(fromStore)
				if err != nil {
					return nil, errors.Errorf("error refreshing token: couldn't parse from []byte: %v", err)
				}

				t, err = typedOptions.RefreshFn(t)
				if err != nil {
					return nil, errors.Errorf("error refreshing token: couldn't refresh: %v", err)
				}

				ret, err := typedOptions.MarshalToStoreFn(t)
				if err != nil {
					return nil, errors.Errorf("error refreshing token: couldn't marshal to []byte: %v", err)
				} else {
					return ret, nil
				}
			}
		}
		if typedOptions.IssueFn != nil {
			plainOptions.IssueFn = func() ([]byte, error) {
				t, err := typedOptions.IssueFn()
				if err != nil {
					return nil, errors.Errorf("error issuing token: couldn't issue: %v", err)
				}

				ret, err := typedOptions.MarshalToStoreFn(t)
				if err != nil {
					return nil, errors.Errorf("error issuing token: couldn't marshal to []byte: %v", err)
				} else {
					return ret, nil
				}
			}
		}

		// Configure plain options
		plainOptions.transformForReturn = func(fromStore []byte) ([]byte, error) {
			t, err := typedOptions.UnmarshalFromStoreFn(fromStore)
			if err != nil {
				return nil, errors.Errorf("error transforming token for return: couldn't parse from []byte: %v", err)
			}

			ret, err := typedOptions.MarshalToReturnFn(t)
			if err != nil {
				return nil, errors.Errorf("error transforming token for return: couldn't marshal to []byte: %v", err)
			} else {
				return ret, nil
			}
		}
		plainOptions.BaseTokenOptions = typedOptions.BaseTokenOptions
	}), nil
}

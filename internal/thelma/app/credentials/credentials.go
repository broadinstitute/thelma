package credentials

import "github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"

type Credentials interface {
	// NewToken returns a new TokenProvider for the given key
	NewTokenProvider(key string, opts ...TokenOption) TokenProvider
}

type credentials struct {
	defaultStore stores.Store
}

func New(credentialsDir string) (Credentials, error) {
	s, err := stores.NewDirectoryStore(credentialsDir)
	if err != nil {
		return nil, err
	}
	return credentials{
		defaultStore: s,
	}, nil
}

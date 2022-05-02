package credentials

import "github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"

type Credentials interface {
	// NewTokenProvider returns a new TokenProvider for the given key
	NewTokenProvider(key string, opts ...TokenOption) TokenProvider
}

type credentials struct {
	defaultStore stores.Store
}

// New returns a new Credentials instance using a directory store rooted at credentialsDir
func New(credentialsDir string) (Credentials, error) {
	s, err := stores.NewDirectoryStore(credentialsDir)
	if err != nil {
		return nil, err
	}
	return credentials{
		defaultStore: s,
	}, nil
}

// NewWithStore returns a new Credentials instance backed by the given store
func NewWithStore(store stores.Store) Credentials {
	return credentials{
		defaultStore: store,
	}
}

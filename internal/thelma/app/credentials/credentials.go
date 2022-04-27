package credentials

import "github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"

type Credentials interface {
	// NewToken returns a new Token for the given key
	NewToken(key string, opts ...TokenOption) Token
}

type credentials struct {
	store stores.Store
}

func New(credentialsDir string) (Credentials, error) {
	s, err := stores.NewDirectoryStore(credentialsDir)
	if err != nil {
		return nil, err
	}
	return credentials{
		store: s,
	}, nil
}

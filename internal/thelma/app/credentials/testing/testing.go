package testing

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"testing"
)

type FakeCredentials struct {
	store stores.Store
	credentials.Credentials
}

// NewFakeCredentials returns a new FakeCredentials that implements the credentials.Credentials interface,
// suitable for use in testing
func NewFakeCredentials(t *testing.T) (*FakeCredentials, error) {
	store, err := stores.NewDirectoryStore(t.TempDir())
	if err != nil {
		return nil, err
	}
	return &FakeCredentials{
		store:       store,
		Credentials: credentials.NewWithStore(store),
	}, nil
}

// AddToStore adds a new credential to the store
func (f *FakeCredentials) AddToStore(key string, value []byte) error {
	return f.store.Write(key, value)
}

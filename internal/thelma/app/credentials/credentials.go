package credentials

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
)

const configKey = "credentials"

type Credentials interface {
	// NewTokenProvider returns a new TokenProvider for the given key
	NewTokenProvider(key string, opts ...TokenOption) TokenProvider
}

type credentialsConfig struct {
	StoreType string `default:"directory" validate:"oneof=directory inmemory"`
}

type credentials struct {
	defaultStore stores.Store
}

// New returns a new Credentials instance using a directory store rooted at credentialsDir
func New(thelmaConfig config.Config, thelmaRoot root.Root) (Credentials, error) {
	var cfg credentialsConfig

	err := thelmaConfig.Unmarshal(configKey, &cfg)
	if err != nil {
		return nil, err
	}

	var store stores.Store

	switch cfg.StoreType {
	case "directory":
		store, err = stores.NewDirectoryStore(thelmaRoot.CredentialsDir())
		if err != nil {
			return nil, err
		}
	default:
		store = stores.NewMapStore()
	}

	return credentials{
		defaultStore: store,
	}, nil
}

// NewWithStore returns a new Credentials instance backed by the given store
func NewWithStore(store stores.Store) Credentials {
	return credentials{
		defaultStore: store,
	}
}

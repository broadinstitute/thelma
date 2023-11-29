package credentials

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
)

const configKey = "credentials"

type Credentials interface {
	// GetTokenProvider returns a TokenProvider for the given key, applying the given options if a new TokenProvider is
	// created to fulfill the request. If a TokenProvider already exists for the given key, it will be returned.
	// It's important for callers to be consistent with the options they pass in, at least for a given Thelma execution,
	// since the options are only used on the first call for a given key.
	//
	// The caching means that there will only ever be one TokenProvider for a given key, so the TokenProvider's
	// concurrency-safety guarantees will reliably apply (and the caller doesn't need to worry about caching the
	// response of this function).
	GetTokenProvider(key string, opts ...TokenOption) TokenProvider
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

package stores

import (
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"os"
	"path"
)

// NewDirectoryStore returns a credential store that will read and write token values to ~/.thelma/credentials/$key,
// where $key is the token's unique identifier/key.
func NewDirectoryStore(dir string) (Store, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	return dirStore{
		dir: dir,
	}, nil
}

type dirStore struct {
	dir string
}

func (s dirStore) Read(key string) ([]byte, error) {
	content, err := os.ReadFile(s.credentialsFile(key))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (s dirStore) Exists(key string) (bool, error) {
	return utils.FileExists(s.credentialsFile(key))
}

func (s dirStore) Write(key string, credential []byte) error {
	return os.WriteFile(s.credentialsFile(key), credential, 0600)
}

func (s dirStore) credentialsFile(key string) string {
	return path.Join(s.dir, key)
}

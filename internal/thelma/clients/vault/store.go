package vault

import (
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials/stores"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"os"
	"path"
)

const vaultTokenFile = ".vault-token"

// NewVaultTokenStore returns a token store that reads and writes values to the current user's ~/.vault-token file.
// (Note that token keys are ignored.)
func NewVaultTokenStore() stores.Store {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Warn().Msgf("failed to identify user home directory, new Vault tokens will not be saved to ~/%s: %v", vaultTokenFile, err)
		return stores.NewNoopStore()
	}
	return newVaultTokenStore(dir)
}

// BackupToken renames ~/.vault-token to ~/vault-token.bak
func BackupToken() error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not identify user home directory: %v", err)
	}
	tokenFile := path.Join(dir, vaultTokenFile)
	backup := fmt.Sprintf("%s.bak", tokenFile)
	log.Info().Msgf("Renaming %s to %s", tokenFile, backup)
	return os.Rename(tokenFile, backup)
}

// package-private constructor for testing
func newVaultTokenStore(dir string) stores.Store {
	var tokenFile string
	if dir != "" {
		tokenFile = path.Join(dir, vaultTokenFile)
	}
	return vaultStore{
		tokenFile: tokenFile,
	}
}

type vaultStore struct {
	tokenFile string
}

func (s vaultStore) Read(key string) ([]byte, error) {
	data, err := os.ReadFile(s.tokenFile)
	if err != nil {
		return []byte{}, err
	}
	return bytes.TrimSpace(data), nil
}

func (s vaultStore) Exists(_ string) (bool, error) {
	return utils.FileExists(s.tokenFile)
}

func (s vaultStore) Write(key string, credential []byte) error {
	log.Info().Msgf("Saving new token to %s", s.tokenFile)
	return os.WriteFile(s.tokenFile, credential, 0600)
}

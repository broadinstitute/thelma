package google

import (
	"encoding/json"
	"fmt"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

var _serviceAccountKeyVaultCache = map[string][]byte{}

// getServiceAccountKeyFromVault caches key bytes in _serviceAccountKeyVaultCache to avoid DDOS-ing our
// Vault server during BEE seeding, when Thelma rapidly and repeatedly authenticates using the same
// SA keys but different subjects (Google's client libraries don't provide convenient variable-subject
// authentication options).
func getServiceAccountKeyFromVault(client *vaultapi.Client, path string, key string) ([]byte, error) {
	cacheKey := fmt.Sprintf("%s/%s", path, key)
	if cached, ok := _serviceAccountKeyVaultCache[cacheKey]; ok {
		return cached, nil
	}

	secret, err := client.Logical().Read(path)
	if err != nil {
		return nil, errors.Errorf("error reading Vault secret at '%s': %v", path, err)
	} else if secret == nil {
		return nil, errors.Errorf("Vault client returned a nil secret for '%s'", path)
	}

	var jsonBytes []byte
	if key == "" {
		// legacy "splatted" key file JSONs
		jsonBytes, err = json.Marshal(secret.Data)
		if err != nil {
			return nil, errors.Errorf("error parsing 'splatted' Vault secret at '%s': %v", path, err)
		}
	} else {
		value, exists := secret.Data[key]
		if !exists {
			return nil, errors.Errorf("Vault secret at '%s' lacks a key '%s'", path, key)
		}
		asString, ok := value.(string)
		if !ok {
			return nil, errors.Errorf("Vault secret at '%s' in key '%s' can't be parsed to string", path, key)
		}
		jsonBytes = []byte(asString)
	}

	_serviceAccountKeyVaultCache[cacheKey] = jsonBytes
	return jsonBytes, nil
}

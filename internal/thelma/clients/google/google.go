package google

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

const configKey = "google"

// Clients client factory for GCP api clients
type Clients interface {
	// Bucket constructs a new Bucket using Thelma's globally-configured Google authentication options
	Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error)
}

type googleConfig struct {
	Auth struct {
		// Type of authentication to use. One of:
		// * "adc": application default credentials
		// * "vault": load service account key from Vault. Should only be used in CI/CD runs.
		Type  string `default:"adc" one-of:"adc vault-sa"`
		Vault struct {
			Path string `default:"secret/devops/thelma/thelma-ci-sa"`
			Key  string `default:"sa-key.json"`
		}
	}
}

type VaultClientFactory interface {
	Vault() (*vaultapi.Client, error)
}

func New(thelmaConfig config.Config, vaultFactory VaultClientFactory) Clients {
	return &google{
		thelmaConfig: thelmaConfig,
		vaultFactory: vaultFactory,
	}
}

type google struct {
	thelmaConfig config.Config
	vaultFactory VaultClientFactory
}

func (g *google) Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error) {
	clientOpts, err := g.clientOptions()
	if err != nil {
		return nil, err
	}

	options = append(options, bucket.WithClientOptions(clientOpts...))
	return bucket.NewBucket(name, options...)
}

func (g *google) clientOptions() ([]option.ClientOption, error) {
	var options []option.ClientOption

	var cfg googleConfig
	err := g.thelmaConfig.Unmarshal(configKey, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading Google client config: %v", err)
	}

	switch cfg.Auth.Type {
	case "adc":
		log.Debug().Msg("Google clients will use application default credentials")
		// no options to add
	case "vault-sa":
		jsonKey, err := g.readServiceAccountKeyFromVault(cfg)
		log.Debug().Msg("Loaded Google service account key from Vault")
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve service account key from Vault: %v", err)
		}
		options = append(options, option.WithCredentialsJSON(jsonKey))
	default:
		return nil, fmt.Errorf("invalid authentication type: %q", cfg.Auth.Type)
	}

	return options, nil
}

func (g *google) readServiceAccountKeyFromVault(cfg googleConfig) ([]byte, error) {
	path := cfg.Auth.Vault.Path
	key := cfg.Auth.Vault.Key

	log.Debug().Msgf("Google clients will use Vault clients (key %s at %s)", key, path)

	vaultClient, err := g.vaultFactory.Vault()
	if err != nil {
		return nil, fmt.Errorf("error reading Google service account key from Vault: %v", err)
	}
	secret, err := vaultClient.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("error reading Google service account key from Vault: %v", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("error reading Google service account key from Vault: no secret at path %s", path)
	}

	value, exists := secret.Data[key]
	if !exists {
		return nil, fmt.Errorf("error reading Google service account key from Vault: missing key %s at path %s", key, path)
	}
	asString, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("error reading Google service account key from Vault: invalid data for key %s at path %s", key, path)
	}

	return []byte(asString), nil
}

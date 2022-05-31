package google

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/tools/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	oauth2google "golang.org/x/oauth2/google"
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"strings"
)

const configKey = "google"

const broadEmailSuffix = "@broadinstitute.org"

var tokenScopes = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/cloud-platform",
}

// Clients client factory for GCP api clients
type Clients interface {
	// Bucket constructs a new Bucket using Thelma's globally-configured Google authentication options
	Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error)
	// Kubectl returns a new Kubectl instance
	Kubectl() (kubectl.Kubectl, error)
}

type googleConfig struct {
	Auth struct {
		// Type of authentication to use. One of:
		// * "adc": application default credentials
		// * "vault": load service account key from Vault. Should only be used in CI/CD runs.
		Type string `default:"adc" one-of:"adc vault-sa"`
		ADC  struct {
			VerifyBroadEmail bool `default:"true"`
		}
		Vault struct {
			Path string `default:"secret/devops/thelma/thelma-ci-sa"`
			Key  string `default:"sa-key.json"`
		}
	}
}

type VaultClientFactory interface {
	Vault() (*vaultapi.Client, error)
}

func New(thelmaConfig config.Config, thelmaRoot root.Root, shellRunner shell.Runner, vaultFactory VaultClientFactory) Clients {
	return &google{
		thelmaConfig: thelmaConfig,
		thelmaRoot:   thelmaRoot,
		shellRunner:  shellRunner,
		vaultFactory: vaultFactory,
	}
}

type google struct {
	thelmaConfig config.Config
	thelmaRoot   root.Root
	shellRunner  shell.Runner
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

func (g *google) Kubectl() (kubectl.Kubectl, error) {
	clientOpts, err := g.clientOptions()
	if err != nil {
		return nil, err
	}
	gkeClient, err := container.NewClusterManagerClient(context.Background(), clientOpts...)
	if err != nil {
		return nil, err
	}

	tokenSource, err := g.tokenSource()
	if err != nil {
		return nil, err
	}

	return kubectl.NewKubectl(g.shellRunner, g.thelmaRoot, tokenSource, gkeClient)
}

func (g *google) clientOptions() ([]option.ClientOption, error) {
	creds, err := g.oauthCredentials()
	if err != nil {
		return nil, err
	}
	return []option.ClientOption{
		option.WithCredentials(creds),
	}, nil
}

func (g *google) tokenSource() (oauth2.TokenSource, error) {
	creds, err := g.oauthCredentials()
	if err != nil {
		return nil, err
	}
	return creds.TokenSource, nil
}

func (g *google) oauthCredentials() (*oauth2google.Credentials, error) {
	var cfg googleConfig
	err := g.thelmaConfig.Unmarshal(configKey, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading Google client config: %v", err)
	}

	switch cfg.Auth.Type {
	case "adc":
		log.Debug().Msg("Google clients will use application default credentials")
		creds, err := oauth2google.FindDefaultCredentialsWithParams(context.Background(), oauth2google.CredentialsParams{
			Scopes: g.copyTokenScopes(),
		})
		if err != nil {
			return nil, fmt.Errorf("error loading Google Cloud ADC credentials: %v", err)
		}
		if cfg.Auth.ADC.VerifyBroadEmail {
			tokenSource := creds.TokenSource
			if err = g.verifyTokenUsesBroadEmail(context.Background(), tokenSource); err != nil {
				return nil, fmt.Errorf("error verifying Google Cloud credentials: %v", err)
			}
		}
		return creds, nil
	case "vault-sa":
		jsonKey, err := g.readServiceAccountKeyFromVault(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve service account key from Vault: %v", err)
		}
		log.Debug().Msg("Loaded Google service account key from Vault")
		return oauth2google.CredentialsFromJSONWithParams(context.Background(), jsonKey, oauth2google.CredentialsParams{
			Scopes: g.copyTokenScopes(),
		})
	default:
		return nil, fmt.Errorf("invalid authentication type: %q", cfg.Auth.Type)
	}
}

func (g *google) copyTokenScopes() []string {
	var scopes []string
	for _, s := range tokenScopes {
		scopes = append(scopes, s)
	}
	return scopes
}

func (g *google) verifyTokenUsesBroadEmail(ctx context.Context, tokenSource oauth2.TokenSource) error {
	oauth2Service, err := googleoauth.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return err
	}
	info, err := oauth2Service.Userinfo.Get().Do()

	if !strings.HasSuffix(info.Email, broadEmailSuffix) {
		return fmt.Errorf(`
Current email %q does not end with %s! Please run

  gcloud auth login <you>@broadinstitute.org --update-adc

and try re-running this command`, info.Email, broadEmailSuffix)
	}

	return nil
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

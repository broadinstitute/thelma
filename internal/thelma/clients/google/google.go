package google

import (
	container "cloud.google.com/go/container/apiv1"
	"context"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/clients/api"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/terraapi"
	"github.com/broadinstitute/thelma/internal/thelma/tools/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
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
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/devstorage.full_control",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
	"openid",
}

// Clients client factory for GCP api clients
type Clients interface {
	// Bucket constructs a new Bucket using Thelma's globally-configured Google authentication options
	Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error)
	// Kubectl returns a new Kubectl instance
	Kubectl() (kubectl.Kubectl, error)
	// Terra returns a new terraapi.TerraClient instance
	Terra() (terraapi.TerraClient, error)
	// SetSubject allows usage of domain-wide delegation privileges, to authenticate as a user via
	// a different account.
	SetSubject(subject string) Clients
}

type googleConfig struct {
	Auth struct {
		// Type of authentication to use. One of:
		// * "adc": application default credentials
		// * "vault": load service account key from Vault. The default here may only be accessible from CI.
		Type string `default:"adc" one-of:"adc vault-sa"`
		ADC  struct {
			VerifyBroadEmail bool `default:"true"`
		}
		Vault struct {
			Path string `default:"secret/devops/thelma/thelma-ci-sa"`
			// Key can be set to an empty string internally at runtime to consume the entire secret as the SA key
			// (for legacy "splatted" key file JSONs)
			Key string `default:"sa-key.json"`
		}
	}
}

func New(thelmaConfig config.Config, thelmaRoot root.Root, shellRunner shell.Runner, vaultFactory api.VaultFactory) Clients {
	return &google{
		thelmaConfig: thelmaConfig,
		thelmaRoot:   thelmaRoot,
		shellRunner:  shellRunner,
		vaultFactory: vaultFactory,
	}
}

func NewUsingVaultSA(
	thelmaConfig config.Config,
	thelmaRoot root.Root,
	shellRunner shell.Runner,
	vaultFactory api.VaultFactory,
	saVaultPath string,
	saVaultKey string,
) Clients {
	customConfig := &googleConfig{}
	customConfig.Auth.Type = "vault-sa"
	customConfig.Auth.Vault.Path = saVaultPath
	customConfig.Auth.Vault.Key = saVaultKey
	return &google{
		thelmaConfig: thelmaConfig,
		thelmaRoot:   thelmaRoot,
		shellRunner:  shellRunner,
		vaultFactory: vaultFactory,
		customConfig: customConfig,
	}
}

func NewUsingADC(
	thelmaConfig config.Config,
	thelmaRoot root.Root,
	shellRunner shell.Runner,
	vaultFactory api.VaultFactory,
	allowNonBroad bool,
) Clients {
	customConfig := &googleConfig{}
	customConfig.Auth.Type = "adc"
	customConfig.Auth.ADC.VerifyBroadEmail = !allowNonBroad
	return &google{
		thelmaConfig: thelmaConfig,
		thelmaRoot:   thelmaRoot,
		shellRunner:  shellRunner,
		vaultFactory: vaultFactory,
		customConfig: customConfig,
	}
}

type google struct {
	thelmaConfig config.Config
	thelmaRoot   root.Root
	shellRunner  shell.Runner
	vaultFactory api.VaultFactory
	customConfig *googleConfig
	subject      string
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

func (g *google) Terra() (terraapi.TerraClient, error) {
	tokenSource, err := g.tokenSource()
	if err != nil {
		return nil, fmt.Errorf("error obtaining token source: %v", err)
	}
	oauth2Service, err := googleoauth.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("error obtaining google oauth2 service: %v", err)
	}
	info, err := oauth2Service.Userinfo.V2.Me.Get().Do()
	if err != nil {
		return nil, fmt.Errorf("error getting google user info: %v", err)
	}
	log.Debug().Msgf("using Terra API client authenticated as %s", info.Email)
	client := terraapi.NewClient(tokenSource, *info)
	return client, nil
}

func (g *google) SetSubject(subject string) Clients {
	g.subject = subject
	return g
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
	if g.customConfig == nil {
		err := g.thelmaConfig.Unmarshal(configKey, &cfg)
		if err != nil {
			return nil, fmt.Errorf("error reading Google client config: %v", err)
		}
	} else {
		log.Debug().Msg("Using Google client with custom configuration")
		cfg = *g.customConfig
	}

	params := oauth2google.CredentialsParams{
		Scopes: g.copyTokenScopes(),
	}
	if len(g.subject) > 0 {
		log.Debug().Msgf("Using Google client with delegated subject: %s", g.subject)
		params.Subject = g.subject
	}

	switch cfg.Auth.Type {
	case "adc":
		log.Debug().Msg("Google clients will use application default credentials")
		creds, err := oauth2google.FindDefaultCredentialsWithParams(context.Background(), params)
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
		log.Debug().Msgf("Loaded Google service account key from Vault (%s .%s)", cfg.Auth.Vault.Path, cfg.Auth.Vault.Key)
		creds, err := oauth2google.CredentialsFromJSONWithParams(context.Background(), jsonKey, params)
		if err != nil {
			return nil, fmt.Errorf("error loading Google Cloud JSON credentials: %v", err)
		}
		return creds, nil
	default:
		return nil, fmt.Errorf("invalid authentication type: %q", cfg.Auth.Type)
	}
}

func (g *google) copyTokenScopes() []string {
	var scopes []string
	scopes = append(scopes, tokenScopes...)
	return scopes
}

func (g *google) verifyTokenUsesBroadEmail(ctx context.Context, tokenSource oauth2.TokenSource) error {
	oauth2Service, err := googleoauth.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return err
	}
	info, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return err
	}

	if !strings.HasSuffix(info.Email, broadEmailSuffix) {
		return fmt.Errorf(`
Current email %q does not end with %s! Please run

  gcloud auth login <you>@broadinstitute.org --update-adc

and try re-running this command`, info.Email, broadEmailSuffix)
	}

	return nil
}

var _serviceAccountKeyVaultCache = map[string][]byte{}

// readServiceAccountKeyFromVault caches key bytes in _serviceAccountKeyVaultCache to avoid DDOS-ing our
// Vault server during BEE seeding, when Thelma rapidly and repeatedly authenticates using the same
// SA keys but different subjects (Google's client libraries don't provide convenient variable-subject
// authentication options).
func (g *google) readServiceAccountKeyFromVault(cfg googleConfig) ([]byte, error) {
	path := cfg.Auth.Vault.Path
	key := cfg.Auth.Vault.Key

	cacheKey := fmt.Sprintf("%s:%s", path, key)
	if cached, present := _serviceAccountKeyVaultCache[cacheKey]; present {
		return cached, nil
	}

	if key == "" {
		log.Debug().Msgf("Google client will use Vault client (splatted key file at %s)", path)
	} else {
		log.Debug().Msgf("Google client will use Vault client (key %s at %s)", key, path)
	}

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

	if key == "" {
		jsonBytes, err := json.Marshal(secret.Data)
		if err != nil {
			return nil, fmt.Errorf("error parsing 'splatted' Google service account key from Vault: %s caused %v", path, err)
		}
		_serviceAccountKeyVaultCache[cacheKey] = jsonBytes
		return jsonBytes, nil
	} else {
		value, exists := secret.Data[key]
		if !exists {
			return nil, fmt.Errorf("error reading Google service account key from Vault: missing key %s at path %s", key, path)
		}
		asString, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("error reading Google service account key from Vault: invalid data for key %s at path %s", key, path)
		}
		jsonBytes := []byte(asString)
		_serviceAccountKeyVaultCache[cacheKey] = jsonBytes
		return jsonBytes, nil
	}
}

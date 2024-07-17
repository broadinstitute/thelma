package google

import (
	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/sqladmin"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/terraapi"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	oauth2google "golang.org/x/oauth2/google"
	"google.golang.org/api/iamcredentials/v1"
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	googlesqladmin "google.golang.org/api/sqladmin/v1"
	"google.golang.org/api/transport"
	"strings"
)

// Clients factory for GCP api clients
type Clients interface {
	// Bucket constructs a new Bucket using Thelma's globally-configured Google authentication options
	Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error)
	// Terra returns a new terraapi.TerraClient instance
	Terra() (terraapi.TerraClient, error)
	// PubSub returns a new google pubsub client
	PubSub(projectId string) (*pubsub.Client, error)
	// ClusterManager returns a new google container cluster manager client
	ClusterManager() (*container.ClusterManagerClient, error)
	// SqlAdmin returns a new google sql admin client
	SqlAdmin() (sqladmin.Client, error)
	// TokenSource returns an oauth TokenSource for this client factory's configured identity
	TokenSource() (oauth2.TokenSource, error)
	// IdTokenGenerator returns a function suitable to be an issueFn for credentials.TokenProvider
	IdTokenGenerator(audience string, serviceAccountChain ...string) (func() ([]byte, error), error)
}

type Options struct {
	// ConfigSource should be provided if neither OptionForceADC nor OptionForceVaultSA are passed.
	ConfigSource config.Config
	// VaultFactory will be lazily executed if Vault SA auth is used.
	VaultFactory func() (*vaultapi.Client, error)
	// Subject, when set, indicates that delegation to that subject should be attempted.
	Subject string

	// configFns is set internally by OptionForceVaultSA and OptionForceADC.
	configFns []func(*googleConfig)
}

type Option func(*Options)

const configKey = "google"

const broadEmailSuffix = "@broadinstitute.org"
const serviceAccountEmailSuffix = ".iam.gserviceaccount.com"
const serviceAccountResourceNamePrefix = "projects/-/serviceAccounts/"

var tokenScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/devstorage.full_control",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
	"openid",
}

type googleConfig struct {
	Auth struct {
		// Type of authentication to use. One of:
		// * "adc": application default credentials
		// * "vault-sa": load service account key from Vault. The default here may only be accessible from CI.
		// * "sa-key": auth using service account key JSON
		Type string `default:"adc" one-of:"adc vault-sa sa-key"`
		ADC  struct {
			VerifyBroadEmail bool `default:"true"`
		}
		Vault struct {
			Path string `default:"secret/devops/thelma/thelma-ci-sa"`
			// Key can be set to an empty string internally at runtime to consume the entire secret as the SA key
			// (for legacy "splatted" key file JSONs)
			Key string `default:"sa-key.json"`
		}
		ServiceAccountKey struct {
			JSON []byte
		}
	}
	TransportLogging struct {
		Enabled bool `default:"false"`
	}
}

func OptionForceVaultSA(
	saVaultPath string,
	saVaultKey string,
) Option {
	return func(o *Options) {
		o.configFns = append(o.configFns, func(c *googleConfig) {
			c.Auth.Type = "vault-sa"
			c.Auth.Vault.Path = saVaultPath
			c.Auth.Vault.Key = saVaultKey
		})
	}
}

func OptionForceSAKey(
	saKeyJson []byte,
) Option {
	return func(o *Options) {
		o.configFns = append(o.configFns, func(c *googleConfig) {
			c.Auth.Type = "sa-key"
			c.Auth.ServiceAccountKey.JSON = saKeyJson
		})
	}
}

func OptionForceADC(
	allowNonBroad bool,
) Option {
	return func(o *Options) {
		o.configFns = append(o.configFns, func(c *googleConfig) {
			c.Auth.Type = "adc"
			c.Auth.ADC.VerifyBroadEmail = !allowNonBroad
		})
	}
}

func New(options ...Option) Clients {
	opts := Options{}
	for _, opt := range options {
		opt(&opts)
	}

	var cfg googleConfig
	if opts.ConfigSource != nil {
		if err := opts.ConfigSource.Unmarshal(configKey, &cfg); err != nil {
			log.Fatal().Err(err).Msgf("failed to unmarshal %s config", configKey)
		}
	}
	for _, fn := range opts.configFns {
		fn(&cfg)
	}

	return &clientsImpl{
		options: opts,
		cfg:     cfg,
	}
}

type clientsImpl struct {
	options Options
	cfg     googleConfig

	cachedGoogleUserinfo    *googleoauth.Userinfo
	cachedGoogleCredentials *oauth2google.Credentials
}

func (c *clientsImpl) googleCredentials() (*oauth2google.Credentials, error) {
	if c.cachedGoogleCredentials == nil {
		params := oauth2google.CredentialsParams{
			Subject: c.options.Subject,
			Scopes:  tokenScopes,
		}
		log.Trace().
			Str("authType", c.cfg.Auth.Type).
			Str("delegationSubject", params.Subject).
			Msg("initializing google credentials")
		switch c.cfg.Auth.Type {
		case "adc":
			var err error
			c.cachedGoogleCredentials, err = oauth2google.FindDefaultCredentialsWithParams(context.Background(), params)
			if err != nil {
				return nil, errors.Errorf("error finding default credentials: %v", err)
			}
			if c.cfg.Auth.ADC.VerifyBroadEmail {
				userinfo, err := c.GoogleUserinfo()
				if err != nil {
					return nil, err
				}
				if !strings.HasSuffix(userinfo.Email, broadEmailSuffix) {
					return nil, errors.Errorf(`
Current email %q does not end with %s! Please run

  gcloud auth login <you>@broadinstitute.org --update-adc

and try re-running this command`, userinfo.Email, broadEmailSuffix)
				}
			}
		case "vault-sa":
			if c.options.VaultFactory == nil {
				return nil, errors.Errorf("vault-sa auth requested but no VaultFactory provided")
			}
			vaultClient, err := c.options.VaultFactory()
			if err != nil {
				return nil, errors.Errorf("error initializing vault client: %v", err)
			}
			jsonBytes, err := getServiceAccountKeyFromVault(vaultClient, c.cfg.Auth.Vault.Path, c.cfg.Auth.Vault.Key)
			if err != nil {
				return nil, errors.Errorf("unable to get key from vault: %v", err)
			}
			c.cachedGoogleCredentials, err = oauth2google.CredentialsFromJSONWithParams(context.Background(), jsonBytes, params)
			if err != nil {
				return nil, errors.Errorf("error initializing credentials from key file from vault: %v", err)
			}
		case "sa-key":
			var err error
			c.cachedGoogleCredentials, err = oauth2google.CredentialsFromJSONWithParams(context.Background(), c.cfg.Auth.ServiceAccountKey.JSON, params)
			if err != nil {
				return nil, errors.Errorf("error initializing credentials from service account key json: %v", err)
			}
		default:
			return nil, errors.Errorf("unknown google auth type %q", c.cfg.Auth.Type)
		}
	}
	return c.cachedGoogleCredentials, nil
}

func (c *clientsImpl) googleClientOptions(usesGrpc bool) ([]option.ClientOption, error) {
	credentials, err := c.googleCredentials()
	if err != nil {
		return nil, errors.Errorf("error initializing google credentials: %v", err)
	}
	if c.cfg.TransportLogging.Enabled && !usesGrpc {
		client, _, err := transport.NewHTTPClient(context.Background(), option.WithCredentials(credentials))
		if err != nil {
			return nil, err
		}
		client.Transport = &loggingTransport{client.Transport}
		return []option.ClientOption{option.WithHTTPClient(client)}, nil
	} else {
		return []option.ClientOption{option.WithCredentials(credentials)}, nil
	}
}

func (c *clientsImpl) GoogleUserinfo() (*googleoauth.Userinfo, error) {
	if c.cachedGoogleUserinfo == nil {
		credentials, err := c.googleCredentials()
		if err != nil {
			return nil, err
		}
		// You might think we could use option.WithCredentials(c.googleClientOptions(false)) here,
		// but it seems not -- ADC human user credentials seem to get messed up if you do that.
		// You'll get an error complaining about "user must be authenticated when user project is
		// provided", which doesn't make much sense but the project would be in the credentials
		// but not the credentials.TokenSource.
		oauth2Service, err := googleoauth.NewService(context.Background(), option.WithTokenSource(credentials.TokenSource))
		if err != nil {
			return nil, errors.Errorf("error initializing oauth2 service: %v", err)
		}
		userinfo, err := googleoauth.NewUserinfoService(oauth2Service).V2.Me.Get().Do()
		if err != nil {
			return nil, errors.Errorf("error connecting to userinfo service: %v", err)
		} else {
			c.cachedGoogleUserinfo = userinfo
		}
	}
	return c.cachedGoogleUserinfo, nil
}

func (c *clientsImpl) TokenSource() (oauth2.TokenSource, error) {
	credentials, err := c.googleCredentials()
	if err != nil {
		return nil, err
	}
	return credentials.TokenSource, nil
}

func (c *clientsImpl) Bucket(name string, options ...bucket.BucketOption) (bucket.Bucket, error) {
	clientOptions, err := c.googleClientOptions(false)
	if err != nil {
		return nil, err
	}
	options = append(options, bucket.WithClientOptions(clientOptions...))
	b, err := bucket.NewBucket(name, options...)
	if err != nil {
		return nil, errors.Errorf("error initializing client library for bucket %q: %v", name, err)
	} else {
		return b, nil
	}
}

func (c *clientsImpl) Terra() (terraapi.TerraClient, error) {
	tokenSource, err := c.TokenSource()
	if err != nil {
		return nil, err
	}
	userinfo, err := c.GoogleUserinfo()
	if err != nil {
		return nil, err
	}
	return terraapi.NewClient(tokenSource, userinfo), nil
}

func (c *clientsImpl) PubSub(projectId string) (*pubsub.Client, error) {
	clientOptions, err := c.googleClientOptions(true)
	if err != nil {
		return nil, err
	}
	return pubsub.NewClient(context.Background(), projectId, clientOptions...)
}

func (c *clientsImpl) ClusterManager() (*container.ClusterManagerClient, error) {
	clientOptions, err := c.googleClientOptions(true)
	if err != nil {
		return nil, err
	}
	return container.NewClusterManagerClient(context.Background(), clientOptions...)
}

func (c *clientsImpl) SqlAdmin() (sqladmin.Client, error) {
	clientOptions, err := c.googleClientOptions(true)
	if err != nil {
		return nil, err
	}
	client, err := googlesqladmin.NewService(context.Background(), clientOptions...)
	if err != nil {
		return nil, err
	}
	return sqladmin.New(client), nil
}

// IdTokenGenerator returns a function to generate ID tokens of a service account for a
// given audience.
//
// When no serviceAccountChain is provided, the authenticated identity will be used
// (an error will be returned if the identity is not a service account).
// When a serviceAccountChain is provided, the last entry will be the name passed
// in the request, with each one before that in order being given as the delegate
// chain.
// See https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/generateIdToken
func (c *clientsImpl) IdTokenGenerator(audience string, serviceAccountChain ...string) (func() ([]byte, error), error) {
	if len(serviceAccountChain) == 0 {
		log.Trace().Msg("IdTokenGenerator called with no service account chain; using authenticated identity")
		userinfo, err := c.GoogleUserinfo()
		if err != nil {
			return nil, err
		}
		serviceAccountChain = []string{userinfo.Email}
	}
	for i := 0; i < len(serviceAccountChain); i++ {
		if !strings.HasSuffix(serviceAccountChain[i], serviceAccountEmailSuffix) {
			return nil, errors.Errorf("IdTokenGenerator called with non-service-account email %q", serviceAccountChain[i])
		} else if !strings.HasPrefix(serviceAccountChain[i], serviceAccountResourceNamePrefix) {
			// Both the name and the delegates need to be in "resource name" format,
			// so we add that prefix if it is missing
			serviceAccountChain[i] = serviceAccountResourceNamePrefix + serviceAccountChain[i]
		}
	}

	clientOptions, err := c.googleClientOptions(false)
	if err != nil {
		return nil, err
	}
	iamcredentialsService, err := iamcredentials.NewService(context.Background(), clientOptions...)
	if err != nil {
		return nil, errors.Errorf("error initializing iamcredentials service: %v", err)
	}
	serviceAccountService := iamcredentials.NewProjectsServiceAccountsService(iamcredentialsService)
	idTokenRequest := &iamcredentials.GenerateIdTokenRequest{
		Audience:     audience,
		IncludeEmail: true,
		Delegates:    serviceAccountChain[1:],
	}
	requestJson, err := idTokenRequest.MarshalJSON()
	if err != nil {
		log.Warn().Err(err).Msg("error marshaling ID token request to JSON")
	} else {
		log.Trace().RawJSON("request", requestJson).Str("name", serviceAccountChain[0]).Msg("performing ID token request")
	}

	return func() ([]byte, error) {
		resp, err := serviceAccountService.GenerateIdToken(serviceAccountChain[0], idTokenRequest).Do()
		if err != nil {
			if userinfo, err2 := c.GoogleUserinfo(); err2 != nil {
				return nil, errors.Errorf("error generating ID token for %s, and couldn't identify caller (GoogleUserinfo() = %v): %v", serviceAccountChain[0], err2, err)
			} else {
				return nil, errors.Errorf("error generating ID token for %s (called by %s): %v", serviceAccountChain[0], userinfo.Email, err)
			}
		}
		return []byte(resp.Token), nil
	}, nil
}

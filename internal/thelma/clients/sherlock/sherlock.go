package sherlock

import (
	"github.com/go-openapi/runtime"
	"net/url"
	"os"
	"strings"

	"github.com/broadinstitute/sherlock/sherlock-go-client/client"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/misc"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const configKey = "sherlock"

// This environment variable is more similar to the Vault client's VAULT_ROLE_ID and VAULT_SECRET_ID environment
// variables than it is to normal config. We only expose this through the environment to help protect against
// this secret value ever being written to a file.
const githubActionsOidcTokenEnvVar = "SHERLOCK_GHA_OIDC_TOKEN"
const sherlockGithubActionsOidcHeader = "X-GHA-OIDC-JWT"

// StateReadWriter is an interface representing the ability to both read and
// create/update thelma's internal state using a sherlock client
type StateReadWriter interface {
	StateLoader
	StateWriter
}

// Client contains an API client for a remote sherlock server. It implements StateReadWriter.
type Client struct {
	client *client.Sherlock
}

type sherlockConfig struct {
	Addr string `default:"https://sherlock.dsp-devops.broadinstitute.org"`
}

// New configures a new Client instance which confers the ability to issue requests against the API of a sherlock server
func New(config config.Config, iapToken string) (*Client, error) {
	sherlockConfig, err := loadConfig(config)
	if err != nil {
		return nil, err
	}

	return configureClientRuntime(sherlockConfig.Addr, iapToken)
}

// NewWithHostnameOverride enables thelma commands to utilize a sherlock client that targets
// a different sherlock instance from the one used for state loading
func NewWithHostnameOverride(addr, iapToken string) (*Client, error) {
	return configureClientRuntime(addr, iapToken)
}

func loadConfig(thelmaConfig config.Config) (sherlockConfig, error) {
	var cfg sherlockConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// sherlock client lib expects host and scheme as separate input values but
// specifying a fqdn in config makes more sense so this helper exists to extract the
// component parts
func extractSchemeAndHost(addr string) (string, string, error) {
	sherlockURL, err := url.Parse(addr)
	if err != nil {
		return "", "", err
	}

	var sherlockHost string
	sherlockHost = sherlockURL.Hostname()

	// account for mock servers via httptest which are assigned a random port on localhost
	if sherlockURL.Port() != "" {
		sherlockHost = strings.Join([]string{sherlockHost, sherlockURL.Port()}, ":")
	}

	return sherlockHost, sherlockURL.Scheme, nil
}

func configureClientRuntime(addr, token string) (*Client, error) {
	hostname, scheme, err := extractSchemeAndHost(addr)
	if err != nil {
		return nil, err
	}

	// setup runtime for openapi client
	transport := httptransport.New(hostname, "", []string{scheme})
	authMechanisms := []runtime.ClientAuthInfoWriter{httptransport.BearerToken(token)}
	if ghaOidcToken := os.Getenv(githubActionsOidcTokenEnvVar); ghaOidcToken != "" {
		authMechanisms = append(authMechanisms, httptransport.APIKeyAuth(sherlockGithubActionsOidcHeader, "header", ghaOidcToken))
	}
	transport.DefaultAuthentication = httptransport.Compose(authMechanisms...)

	sherlockClient := client.New(transport, strfmt.Default)
	return &Client{client: sherlockClient}, nil
}

// getStatus is used in tests to verify that an initialzied sherlock client
// can successfully issue a request against a remote sherlock backend
func (c *Client) getStatus() error {
	params := misc.NewGetStatusParams()
	_, err := c.client.Misc.GetStatus(params)
	return err
}

package sherlock

import (
	"net/url"
	"strings"

	"github.com/broadinstitute/sherlock/clients/go/client"
	"github.com/broadinstitute/sherlock/clients/go/client/misc"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const configKey = "sherlock"

// Client contains an API client for a remote sherlock server
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

// NewWithHostNameOverride enables thelma commands to utilize a sherlock client that targets
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
// specifying an fqdn in config makes more sense so this helper exists to extract the
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
	transport.DefaultAuthentication = httptransport.BearerToken(token)

	sherlockClient := client.New(transport, strfmt.Default)
	client := &Client{client: sherlockClient}
	return client, nil
}

// getStatus is used in tests to verify that an initialzied sherlock client
// can successfully issue a request against a remote sherlock backend
func (c *Client) getStatus() error {
	params := misc.NewGetStatusParams()
	_, err := c.client.Misc.GetStatus(params)
	return err
}

package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"net/url"
	"strings"

	"github.com/broadinstitute/sherlock/sherlock-go-client/client"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/misc"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const configKey = "sherlock"

// Client is an interface representing the ability to both read and
// create/update thelma's internal state using a sherlock client. We have this
// interface largely so we can generate a mock for it.
type Client interface {
	StateLoader
	StateWriter
	ChartVersionUpdater
	GetStatus() error
}

type Options struct {
	Addr                 string
	ConfigSource         config.Config
	GhaOidcTokenProvider credentials.TokenProvider
}

type Option func(*Options)

type sherlockConfig struct {
	Addr string `default:"https://sherlock.dsp-devops.broadinstitute.org"`
}

func NewClient(iapToken string, options ...Option) (Client, error) {
	opts := &Options{}
	for _, option := range options {
		option(opts)
	}

	if opts.ConfigSource != nil {
		var cfg sherlockConfig
		if err := opts.ConfigSource.Unmarshal(configKey, &cfg); err != nil {
			return nil, err
		}

		if opts.Addr == "" {
			opts.Addr = cfg.Addr
		}
	}

	hostname, scheme, err := extractSchemeAndHost(opts.Addr)
	if err != nil {
		return nil, err
	}

	// setup runtime for openapi client
	transport := httptransport.New(hostname, "", []string{scheme})
	transport.DefaultAuthentication = makeClientAuthWriter(iapToken, opts.GhaOidcTokenProvider)

	return &clientImpl{client: client.New(transport, strfmt.Default)}, nil
}

// clientImpl contains an API client for a remote sherlock server. It implements Client.
type clientImpl struct {
	client *client.Sherlock
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

// GetStatus is used in tests to verify that an initialized Client
// can successfully issue a request against a remote sherlock backend
func (c *clientImpl) GetStatus() error {
	params := misc.NewGetStatusParams()
	_, err := c.client.Misc.GetStatus(params)
	return err
}

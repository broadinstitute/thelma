package sherlock

import (
	"github.com/broadinstitute/sherlock/clients/go/client"
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
	Addr   string `default:"sherlock.dsp-devops.broadinstitute.org"`
	Scheme string `default:"https"`
}

// New configures a new Client instance which confers the ability to issue requests against the API of a sherlock server
func New(config config.Config, iapToken string) (*Client, error) {

	sherlockConfig, err := loadConfig(config)
	if err != nil {
		return nil, err
	}

	// setup runtime for openapi client
	transport := httptransport.New(sherlockConfig.Addr, "", []string{sherlockConfig.Scheme})
	transport.DefaultAuthentication = httptransport.BearerToken(iapToken)

	sherlockClient := client.New(transport, strfmt.Default)

	sherlock := &Client{client: sherlockClient}
	return sherlock, nil
}

func loadConfig(thelmaConfig config.Config) (sherlockConfig, error) {
	var cfg sherlockConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

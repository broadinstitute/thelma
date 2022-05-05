// Package clients contains convenience constructors API clients that Thelma uses
package clients

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/clients/iap"
	"github.com/broadinstitute/thelma/internal/thelma/clients/vault"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
)

// Clients convenience builders for client objects used in Thelma commands
type Clients interface {
	// IAPToken returns a valid dsp-tools-k8s IAP token (as a string), or an error
	IAPToken() (string, error)
	// Vault returns a Vault client for the DSP Vault instance
	Vault() (*vaultapi.Client, error)
	// ArgoCD returns a client for the DSP ArgoCD instance
	ArgoCD() (argocd.ArgoCD, error)
}

func New(config config.Config, creds credentials.Credentials, runner shell.Runner) Clients {
	return &clients{
		config: config,
		creds:  creds,
		runner: runner,
	}
}

type clients struct {
	config config.Config
	creds  credentials.Credentials
	runner shell.Runner
}

func (c *clients) IAPToken() (string, error) {
	vaultClient, err := c.Vault()
	if err != nil {
		return "", err
	}

	tokenProvider, err := iap.TokenProvider(c.config, c.creds, vaultClient, c.runner)
	if err != nil {
		return "", err
	}

	token, err := tokenProvider.Get()
	if err != nil {
		return "", err
	}

	return string(token), nil
}

func (c *clients) Vault() (*vaultapi.Client, error) {
	return vault.NewClient(c.config, c.creds)
}

func (c *clients) ArgoCD() (argocd.ArgoCD, error) {
	iapToken, err := c.IAPToken()
	if err != nil {
		return nil, err
	}

	vaultClient, err := c.Vault()
	if err != nil {
		return nil, err
	}

	return argocd.New(c.config, c.runner, iapToken, vaultClient)
}

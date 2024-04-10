// Package clients contains convenience constructors API clients that Thelma uses
package clients

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/github/gha"
	"github.com/broadinstitute/thelma/internal/thelma/clients/newiap"
	"sync"

	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/clients/github"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/clients/iap"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/clients/slack"
	"github.com/broadinstitute/thelma/internal/thelma/clients/vault"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
)

// Clients convenience builders for client objects used in Thelma commands
type Clients interface {
	// IAP returns a credentials.TokenProvider for dsp-tools-k8s IAP tokens.
	IAP() (credentials.TokenProvider, error)
	// NewIAP is like IAP but for dsp-devops-super-prod.
	NewIAP() (credentials.TokenProvider, error)
	// IAPToken returns a valid dsp-tools-k8s IAP token (as a string), or an error.
	// This is a convenience method on top of IAP().
	IAPToken() (string, error)
	// NewIAPToken is like IAPToken but for dsp-devops-super-prod.
	NewIAPToken() (string, error)
	// Vault returns a Vault client for the DSP Vault instance
	Vault() (*vaultapi.Client, error)
	// ArgoCD returns a client for the DSP ArgoCD instance
	ArgoCD() (argocd.ArgoCD, error)
	// Github returns a wrapper around a github api client instance
	Github() (*github.Client, error)
	// Google returns a client factory for GCP clients, using Thelma's default configuration
	Google(options ...google.Option) google.Clients
	// Kubernetes returns a factory for Kubernetes clients
	Kubernetes() kubernetes.Clients
	// Sherlock returns a swagger API client for a sherlock server instance
	Sherlock(options ...sherlock.Option) (sherlock.Client, error)
	// Slack returns a wrapper around the official API client
	Slack() (slack.Slack, error)
}

func New(thelmaConfig config.Config, thelmaRoot root.Root, creds credentials.Credentials, runner shell.Runner) (Clients, error) {
	return &clients{
		thelmaConfig: thelmaConfig,
		thelmaRoot:   thelmaRoot,
		creds:        creds,
		runner:       runner,
	}, nil
}

type clients struct {
	thelmaConfig config.Config
	thelmaRoot   root.Root
	creds        credentials.Credentials
	runner       shell.Runner
	mutex        sync.Mutex
	kubernetes   kubernetes.Clients
}

func (c *clients) Google(options ...google.Option) google.Clients {
	opts := append([]google.Option{
		func(o *google.Options) {
			o.ConfigSource = c.thelmaConfig
			o.VaultFactory = c.Vault
		},
	}, options...)
	return google.New(opts...)
}

func (c *clients) IAP() (credentials.TokenProvider, error) {
	return iap.TokenProvider(c.thelmaConfig, c.creds, c.Vault, c.Google, c.runner)
}

func (c *clients) NewIAP() (credentials.TokenProvider, error) {
	return newiap.TokenProvider(c.thelmaConfig, c.creds, c.Google, c.runner)
}

func (c *clients) IAPToken() (string, error) {
	if tokenProvider, err := c.IAP(); err != nil {
		return "", err
	} else if token, err := tokenProvider.Get(); err != nil {
		return "", err
	} else {
		return string(token), nil
	}
}

func (c *clients) NewIAPToken() (string, error) {
	if tokenProvider, err := c.NewIAP(); err != nil {
		return "", err
	} else if token, err := tokenProvider.Get(); err != nil {
		return "", err
	} else {
		return string(token), nil
	}
}

func (c *clients) Vault() (*vaultapi.Client, error) {
	return vault.NewClient(c.thelmaConfig, c.creds)
}

func (c *clients) ArgoCD() (argocd.ArgoCD, error) {
	iapToken, err := c.NewIAPToken()
	if err != nil {
		return nil, err
	}

	return argocd.New(c.thelmaConfig, c.runner, iapToken, c.Vault)
}

func (c *clients) Sherlock(options ...sherlock.Option) (sherlock.Client, error) {
	iapProvider, err := c.IAP()
	if err != nil {
		return nil, err
	}
	ghaOidcProvider, err := gha.NewGhaOidcProvider(c.thelmaConfig, c.creds)
	if err != nil {
		return nil, err
	}
	opts := []sherlock.Option{
		func(options *sherlock.Options) {
			options.ConfigSource = c.thelmaConfig
			options.IapTokenProvider = iapProvider
			options.GhaOidcTokenProvider = ghaOidcProvider
		},
	}
	opts = append(opts, options...)
	return sherlock.NewClient(opts...)
}

func (c *clients) Kubernetes() kubernetes.Clients {
	// the Kubernetes client factory does some caching, so we
	// lazy-initialize and cache a single instance
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.kubernetes != nil {
		return c.kubernetes
	}
	_kubernetes := kubernetes.New(c.thelmaRoot, c.runner, c.Google())
	c.kubernetes = _kubernetes
	return c.kubernetes
}

func (c *clients) Slack() (slack.Slack, error) {
	return slack.New(c.thelmaConfig, c.Vault)
}

func (c *clients) Github() (*github.Client, error) {
	return github.New(github.WithDefaults(c.thelmaConfig, c.creds, c.Vault))
}

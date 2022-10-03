// Package clients contains convenience constructors API clients that Thelma uses
package clients

import (
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/clients/iap"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	slackapi "github.com/broadinstitute/thelma/internal/thelma/clients/slack"
	"github.com/broadinstitute/thelma/internal/thelma/clients/vault"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
	"os"
)

// Clients convenience builders for client objects used in Thelma commands
type Clients interface {
	// IAPToken returns a valid dsp-tools-k8s IAP token (as a string), or an error
	IAPToken() (string, error)
	// Vault returns a Vault client for the DSP Vault instance
	Vault() (*vaultapi.Client, error)
	// ArgoCD returns a client for the DSP ArgoCD instance
	ArgoCD() (argocd.ArgoCD, error)
	// Google returns a client factory for GCP clients, using Thelma's default configuration
	Google() google.Clients
	// GoogleUsingVaultSA is like Google but allows a vault path/key for the service account key file
	// to be specified directly at runtime
	GoogleUsingVaultSA(string, string) google.Clients
	// GoogleUsingADC is like Google but forces usage of the environment's Application Default Credentials,
	// optionally allowing non-Broad email addresses
	GoogleUsingADC(bool) google.Clients
	// Sherlock returns a swagger API client for a sherlock server instance
	Sherlock() (*sherlock.Client, error)
	// Slack for the DSP Slack instance
	Slack() (*slackapi.SlackAPI, error)
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
}

func (c *clients) Google() google.Clients {
	return google.New(c.thelmaConfig, c.thelmaRoot, c.runner, c)
}

func (c *clients) GoogleUsingVaultSA(vaultPath string, vaultKey string) google.Clients {
	return google.NewUsingVaultSA(c.thelmaConfig, c.thelmaRoot, c.runner, c, vaultPath, vaultKey)
}

func (c *clients) GoogleUsingADC(allowNonBroad bool) google.Clients {
	return google.NewUsingADC(c.thelmaConfig, c.thelmaRoot, c.runner, c, allowNonBroad)
}

func (c *clients) IAPToken() (string, error) {
	vaultClient, err := c.Vault()
	if err != nil {
		return "", err
	}

	tokenProvider, err := iap.TokenProvider(c.thelmaConfig, c.creds, vaultClient, c.runner)
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
	return vault.NewClient(c.thelmaConfig, c.creds)
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

	return argocd.New(c.thelmaConfig, c.runner, iapToken, vaultClient)
}

func (c *clients) Sherlock() (*sherlock.Client, error) {
	iapToken, err := c.IAPToken()
	if err != nil {
		return nil, err
	}
	return sherlock.New(c.thelmaConfig, iapToken)
}

func (c *clients) Slack() (*slackapi.SlackAPI, error) {
	var token *bytes.Buffer
	envToken := new(bytes.Buffer)
	len, _ := envToken.WriteString(os.Getenv("SLACK_TOKEN"))
	if len == 0 {
		cloudToken, err := slackapi.GetSlackToken()
		if err != nil {
			return nil, fmt.Errorf("no SLACK_TOKEN and retrieval failed with %v", err)
		}
		token = cloudToken
	} else {
		token = envToken
	}
	return slackapi.New(token.String())
}

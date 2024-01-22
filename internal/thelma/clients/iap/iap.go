package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

//
// Authors:
// * Jack Warren: design & proof-of-concept
// * Chelsea Hoover: massaged into thelma
//

const (
	// configKey prefix used for configuration for this package
	configKey = "iap"
	// tokenKey unique name for IAP tokens issued by this package, used to identify it in Thelma's token storage
	tokenKey = "iap-oauth-token"
	// The default that credentials.TokenOptions would set up based on
	defaultTokenEnvVar = "IAP_OAUTH_TOKEN"
	// The old value that this package accepted directly, short-circuiting the Vault logic. The actual value used
	// was "THELMA_IAP_ID_TOKEN", but credentials.TokenProvider will automatically check for "THELMA_" prefix too.
	backwardsCompatibilityTokenEnvVar = "IAP_ID_TOKEN"
)

type iapConfig struct {
	Provider string `default:"browser"  validate:"oneof=google workloadidentity browser"`
	// Audience is the client ID of the OAuth credentials, but it isn't secret, and embedding it in Thelma's
	// config helps some Provider implementations avoid needing Vault access.
	Audience         string `default:"1038484894585-k8qvf7l876733laev0lm8kenfa2lj6bn.apps.googleusercontent.com"`
	OAuthCredentials struct {
		VaultPath string `default:"secret/dsp/identity-aware-proxy/dsp-tools-k8s/dsp-tools-k8s-iap-oauth_client-credentials.json"`
		VaultKey  string `default:"web"`
	}
	WorkloadIdentity struct {
		ServiceAccount string `default:"default"` // default to using compute engine default service account
	}
}

// TokenProvider returns a new token provider for IAP tokens
func TokenProvider(
	thelmaConfig config.Config,
	creds credentials.Credentials,
	lazyVaultClient func() (*vaultapi.Client, error),
	lazyGoogleClient func(options ...google.Option) google.Clients,
	runner shell.Runner) (credentials.TokenProvider, error) {
	var cfg iapConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return nil, err
	}

	switch cfg.Provider {
	case "google":
		return googleProvider(creds, cfg, lazyGoogleClient())
	case "workloadidentity":
		return workloadIdentityProvider(creds, cfg), nil
	case "browser":
		if vaultClient, err := lazyVaultClient(); err != nil {
			return nil, errors.Wrap(err, "unable to create Vault client for usage with the 'browser' IAP provider")
		} else {
			return browserProvider(creds, cfg, vaultClient, runner)
		}
	default:
		return nil, errors.Errorf("unknown iap provider type: %s", cfg.Provider)
	}
}

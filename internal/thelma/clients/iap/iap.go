package iap

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
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

	// tokenKey is *not* a constant for this client in the way that it is for many others.
	// The IAP client can be used against different project, so the token key is based on
	// the project ID. See the Project type for more information.
)

type iapConfig struct {
	// Provider is the mechanism to generate an IAP ID token with.
	// - "google" borrows Thelma's Google authentication (see google.Clients; only works for a service account)
	// - "workloadidentity" uses the GCP metadata server (only works for a service account)
	// - "browser" uses the OAuth flow for desktop applications (only works for a human)
	Provider string `default:"browser"  validate:"oneof=google workloadidentity browser"`

	// Projects contains OAuth credentials to serve as an IAP client for different projects. THESE ARE NOT SECRET!!
	// At least, the defaults here aren't. These defaults are "desktop" client credentials that are allowed
	// programmatic access but aren't allowed anything else -- all they do is let the caller prove their identity
	// to our tools. We embed them here because Thelma is a client-side application and we need to bootstrap
	// somehow. These values *will* trip all manner of "don't commit secrets" alarms, but our usage is one of the
	// very rare cases where it's okay.
	//
	// The client ID is used by all providers, as the audience at the very least. The client secret is only used by
	// the "browser" provider to do the desktop-application flow.
	//
	// https://cloud.google.com/iap/docs/authentication-howto#authenticating_from_a_desktop_app
	Projects struct {
		// DspDevopsSuperProd correlates to dsp-devops-super-prod. Again, these OAuth credentials are not secret.
		// Sherlock, Beehive, and ArgoCD all use this project.
		//
		// https://broadinstitute.slack.com/archives/CADU7L0SZ/p1712604883191549
		// https://github.com/broadinstitute/terraform-ap-deployments/tree/master/devops-super-prod#desktop-iap-credentials
		// https://broadworkbench.atlassian.net/browse/DDO-3605
		DspDevopsSuperProd struct {
			ClientID     string `default:"257801540345-1gqi6qi66bjbssbv01horu9243el2r8b.apps.googleusercontent.com"` // Intentionally public!
			ClientSecret string `default:"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"`                                      // Intentionally public!
		}
		// DspToolsK8s correlates to dsp-tools-k8s. Again, these OAuth credentials are not secret.
		// Dev instances of Sherlock and Beehive use this project, but Thelma's primary contact with this project is
		// the Prometheus push gateway.
		//
		// https://broadworkbench.atlassian.net/browse/DDO-3632
		DspToolsK8s struct {
			ClientID     string `default:"1038484894585-7hklrpn2pfqukro753n5oa335qf8adl4.apps.googleusercontent.com"` // Intentionally public!
			ClientSecret string `default:"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"`                                       // Intentionally public!
		}
	}

	WorkloadIdentity struct {
		// ServiceAccount is the service account to use for the workload identity provider. Note that "default" doesn't
		// necessarily mean the default service account for the project; it means the service account that the
		// individual workload authenticates as by default.
		ServiceAccount string `default:"default"`
	}
}

// TokenProvider returns a new token provider for IAP tokens
func TokenProvider(
	thelmaConfig config.Config,
	creds credentials.Credentials,
	lazyGoogleClient func(options ...google.Option) google.Clients,
	runner shell.Runner,
	project Project,
) (credentials.TokenProvider, error) {
	var cfg iapConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return nil, err
	}

	switch cfg.Provider {
	case "google":
		return googleProvider(creds, cfg, lazyGoogleClient(), project)
	case "workloadidentity":
		return workloadIdentityProvider(creds, cfg, project)
	case "browser":
		return browserProvider(creds, cfg, runner, project)
	default:
		return nil, errors.Errorf("unknown iap provider type: %s", cfg.Provider)
	}
}

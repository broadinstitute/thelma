package newiap

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"time"
)

const (
	// how long to wait before timing out compute engine metadata request
	computeEngineMetadataRequestTimeout = 15 * time.Second
)

func workloadIdentityProvider(creds credentials.Credentials, cfg iapConfig) credentials.TokenProvider {
	return creds.GetTokenProvider(tokenKey, func(options *credentials.TokenOptions) {
		options.EnvVars = []string{defaultTokenEnvVar, backwardsCompatibilityTokenEnvVar}
		options.IssueFn = workloadIdentityIdtokenIssuer(cfg.WorkloadIdentity.ServiceAccount, cfg.ClientID)
		options.ValidateFn = idtokenValidator
	})
}

func workloadIdentityIdtokenIssuer(serviceAccount string, audience string) func() ([]byte, error) {
	metadataUrl := fmt.Sprintf("http://metadata/computeMetadata/v1/instance/service-accounts/%s/identity?audience=%s&format=full",
		serviceAccount, audience)

	return func() ([]byte, error) {
		log.Trace().
			Str("serviceAccount", serviceAccount).
			Str("audience", audience).
			Str("metadataUrl", metadataUrl).
			Msgf("issuing ID token for %s via workload identity", serviceAccount)
		req, err := http.NewRequest(http.MethodGet, metadataUrl, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Metadata-Flavor", "Google")
		client := http.Client{
			Timeout: computeEngineMetadataRequestTimeout,
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("received non-200 response code from compute engine metadata: %v", resp.StatusCode)
		}
		token, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err = resp.Body.Close(); err != nil {
			return nil, err
		}
		return token, nil
	}
}

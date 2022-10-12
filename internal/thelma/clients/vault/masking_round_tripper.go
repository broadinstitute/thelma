package vault

import (
	"bytes"
	"encoding/json"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"regexp"
)

// safeFieldNames a list of field names in Vault secrets to exclude from automatic masking (these
// fields almost always contain non-secret data and lead to useful information being
// masked in Thelma logs)
var safeFieldNames = []string{
	// usually-safe fields in service account key secrets
	"auth_provider_x509_cert_url",
	"auth_uri",
	"client_email",
	"client_x509_cert_url",
	"project_id",
	"type",
	// usually-safe fields in database credential secrets
	"app_sql_user",
	"user",
	"username",
}

const secretsEngineEndpoint = `^/v\d/secret/`

// MaskingRoundTripper implements the http.RoundTripper interface, automatically masking any secrets returned from the Vault API
type MaskingRoundTripper struct {
	inner          http.RoundTripper       // inner round tripper this one delegates to (this is what actually makes the request)
	maskFn         func(secrets ...string) // maskFn custom masking function (should only be used in unit tests -- by default we use logging.MaskSecret)
	safeFieldNames set.StringSet           // safeFieldNames set of field names in secrets for which values should not automatically be masked
}

func newMaskingRoundTripper(inner http.RoundTripper) MaskingRoundTripper {
	return MaskingRoundTripper{
		inner:          inner,
		maskFn:         logging.MaskSecret,
		safeFieldNames: set.NewStringSet(safeFieldNames...),
	}
}

func (m MaskingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := log.With().Str("method", req.Method).Str("path", req.URL.Path).Logger()

	// Send the request, get the response (or the error)
	resp, err := m.inner.RoundTrip(req)

	if err != nil {
		return resp, err
	}

	// If we got a non-2xx status code, return response without attempting to auto-mask
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Debug().Msg("received non-2xx response from Vault server, won't attempt to mask")
		return resp, err
	}

	// Read response body
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("error reading response from Vault")
		return resp, err
	}

	// Copy response body to a fresh io.Reader so the Vault client can still read it when we're done
	resp.Body = io.NopCloser(bytes.NewReader(content))

	// Deserialize response body
	var secret vaultapi.Secret
	if err := json.Unmarshal(content, &secret); err != nil {
		// there may be some Vault API calls that don't neatly deserialize into a Secret; log a warning & move on
		logger.Warn().Err(err).Msg("error unmarshalling response from Vault")
		return resp, nil
	}

	// If the response includes a client token, mask it
	if secret.Auth != nil && len(secret.Auth.ClientToken) > 0 {
		m.maskFn(secret.Auth.ClientToken)
		logger.Debug().Msgf("automatically masked Vault token")
	}

	// If this request was to the secrets engine API and the response included data,
	// mask every string value in the response.
	if regexp.MustCompile(secretsEngineEndpoint).MatchString(req.URL.Path) && len(secret.Data) > 0 {
		count := 0
		for field, value := range secret.Data {
			if m.safeFieldNames.Exists(field) {
				continue
			}
			asString, ok := value.(string)
			if ok {
				logger.Debug().Str("field", field).Msgf("masked value in Vault secret")
				m.maskFn(asString)
				count++
			}
		}
	}

	return resp, nil
}

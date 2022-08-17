package vault

import (
	"bytes"
	"encoding/json"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

// MaskingRoundTripper implements the http.RoundTripper interface, automatically masking any secrets returned from the Vault API
type MaskingRoundTripper struct {
	inner  http.RoundTripper       // inner round tripper this one delegates to (this is what actually makes the request)
	maskFn func(secrets ...string) // maskFn custom masking function (should only be used in unit tests -- by default we use logging.MaskSecret)
}

func newMaskingRoundTripper(inner http.RoundTripper) MaskingRoundTripper {
	return MaskingRoundTripper{
		inner:  inner,
		maskFn: logging.MaskSecret,
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
		log.Debug().Msgf("automatically masked Vault token")
	}

	// If the response included data, mask every string value
	if len(secret.Data) > 0 {
		count := 0
		for _, value := range secret.Data {
			asString, ok := value.(string)
			if ok {
				m.maskFn(asString)
				count++
			}
		}
		log.Debug().Msgf("automatically masked %d fields in Vault response", count)
	}

	return resp, nil
}

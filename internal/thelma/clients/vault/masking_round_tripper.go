package vault

import (
	"bytes"
	"encoding/json"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

// MaskingRoundTripper implements the http.RoundTripper interface, automatically masking any secrets returned from the Vault API
type MaskingRoundTripper struct {
	inner  http.RoundTripper
	maskFn func(secrets ...string)
}

func (m MaskingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := log.With().Str("method", req.Method).Str("path", req.URL.Path).Logger()

	// Send the request, get the response (or the error)
	resp, err := m.inner.RoundTrip(req)

	if err != nil {
		return resp, err
	}

	// if non-2xx status code, return response without attempting to auto-mask
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Debug().Msg("received non-2xx response from Vault server, won't attempt to mask")
		return resp, err
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("error reading response from Vault")
		return resp, err
	}

	// make fresh io.Reader with body content and add it to the response body so Vault client can still read it when we're done
	resp.Body = ioutil.NopCloser(bytes.NewReader(content))

	var secret vaultapi.Secret
	if err := json.Unmarshal(content, &secret); err != nil {
		// there may be some Vault API calls that don't neatly deserialize into a Secret; log a warning & move on
		logger.Warn().Err(err).Msg("error unmarshalling response from Vault")
		return resp, nil
	}

	if secret.Auth != nil && len(secret.Auth.ClientToken) > 0 {
		m.maskSecrets(secret.Auth.ClientToken)
		log.Debug().Msgf("automatically masked Vault token")
	}

	// mask every string field in the secret
	if len(secret.Data) > 0 {
		count := 0
		for _, value := range secret.Data {
			asString, ok := value.(string)
			if ok {
				m.maskSecrets(asString)
				count++
			}
		}
		log.Debug().Msgf("automatically masked %d fields in Vault response", count)
	}

	return resp, nil
}

func (m MaskingRoundTripper) maskSecrets(secrets ...string) {
	if m.maskFn != nil {
		m.maskFn(secrets...)
	} else {
		logging.MaskSecret(secrets...)
	}
}

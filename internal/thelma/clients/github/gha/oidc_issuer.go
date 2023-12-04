package gha

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

const (
	ghaOidcRequestTokenEnvVar = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	ghaOidcRequestUrlEnvVar   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	ghaOidcPermissionsDocUrl  = "https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token"
)

func getOidcRequestValues() (_requestUrl, _requestToken string, err error) {
	if requestUrl, requestUrlPresent := os.LookupEnv(ghaOidcRequestUrlEnvVar); !requestUrlPresent {
		err = fmt.Errorf("GitHub Actions did not inject %s into this job, either the job permissions are incorrect or it is being run from a fork (see `id-token` at %s)", ghaOidcRequestUrlEnvVar, ghaOidcPermissionsDocUrl)
		return
	} else if requestUrl == "" {
		err = fmt.Errorf("%s was specifically set to empty", ghaOidcRequestUrlEnvVar)
		return
	} else {
		_requestUrl = requestUrl
	}
	if requestToken, requestTokenPresent := os.LookupEnv(ghaOidcRequestTokenEnvVar); !requestTokenPresent {
		err = fmt.Errorf("GitHub Actions did not inject %s into this job, either the job permissions are incorrect or it is being run from a fork (see `id-token` at %s)", ghaOidcRequestTokenEnvVar, ghaOidcPermissionsDocUrl)
		return
	} else if requestToken == "" {
		err = fmt.Errorf("%s was specifically set to empty", ghaOidcRequestTokenEnvVar)
	} else {
		_requestToken = requestToken
	}
	return
}

func getOidcToken() ([]byte, error) {
	requestUrl, requestToken, err := getOidcRequestValues()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, errors.Errorf("failed to create request to %s: %v", requestUrl, err)
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", requestToken))
	request.Header.Add("Accept", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Errorf("failed to make request to %s: %v", requestUrl, err)
	}
	defer func(readCloser io.ReadCloser) {
		_ = readCloser.Close()
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		log.Warn().Msgf("GHA OIDC request to %s returned status code %d, attempting to continue since printing the response for debugging would be unsafe in CI anyway", requestUrl, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read response body from %s: %v", requestUrl, err)
	}

	// This type isn't really documented but you can infer it based on the example at
	// https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-cloud-providers#requesting-the-jwt-using-environment-variables
	type responseJson struct {
		Value json.RawMessage `json:"value"`
	}
	var result responseJson
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, errors.Errorf("failed to unmarshal response body from %s: %v", requestUrl, err)
	}
	return result.Value, nil
}

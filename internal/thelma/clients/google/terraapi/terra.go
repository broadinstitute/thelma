package terraapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"
)

type TerraClient interface {
	FirecloudOrch(release terra.AppRelease) FirecloudOrchClient
	Sam(release terra.AppRelease) SamClient
	GoogleUserInfo() googleoauth.Userinfo
}

type terraClient struct {
	tokenSource   oauth2.TokenSource
	userInfo      googleoauth.Userinfo
	httpClient    http.Client
	retryAttempts uint
	retryDelay    time.Duration
}

func NewClient(tokenSource oauth2.TokenSource, userInfo googleoauth.Userinfo) TerraClient {
	return &terraClient{tokenSource: tokenSource, userInfo: userInfo, httpClient: http.Client{}}
}

func (c *terraClient) FirecloudOrch(appRelease terra.AppRelease) FirecloudOrchClient {
	return &firecloudOrchClient{
		terraClient: c,
		appRelease:  appRelease,
	}
}

func (c *terraClient) Sam(appRelease terra.AppRelease) SamClient {
	return &samClient{
		terraClient: c,
		appRelease:  appRelease,
	}
}

func (c *terraClient) GoogleUserInfo() googleoauth.Userinfo {
	return c.userInfo
}

func (c *terraClient) doJsonRequest(method string, url string, body io.Reader) (*http.Response, string, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, "", err
	}
	req.Header.Add("content-type", "application/json")
	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, "", err
	}
	token.SetAuthHeader(req)
	response, err := c.httpClient.Do(req)
	if err != nil {
		return response, "", err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(response.Body)
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return response, "", err
	}
	if response.StatusCode > 299 {
		return response, string(responseBody), errors.Errorf("%s from %s (%s)", response.Status, url, responseBody)
	}
	return response, string(responseBody), nil
}

func (c *terraClient) doJsonRequestWithRetries(method string, url string, bodyData interface{}) (*http.Response, string, error) {
	retryAttempts := c.retryAttempts
	if retryAttempts == 0 {
		retryAttempts = defaultRetryAttempts
	}

	retryDelay := c.retryDelay
	if retryDelay == 0 {
		retryDelay = defaultRetryDelay
	}

	var resp *http.Response
	var responseBody string
	var err error

	requestBody, err := json.Marshal(bodyData)
	if err != nil {
		return nil, "", errors.Errorf("error marshalling request body for %s %s: %v", method, url, err)
	}

	requestFn := func() error {
		resp, responseBody, err = c.doJsonRequest(method, url, bytes.NewBuffer(requestBody))
		return err
	}

	var count int
	if retryErr := retry.Do(
		requestFn,
		retry.Attempts(retryAttempts),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(retryDelay),
		retry.OnRetry(func(n uint, err error) {
			count++
			log.Warn().Err(err).Msgf("%s %s failed (attempt %d of %d): %v", method, url, n, defaultRetryAttempts, err)
		}),
		retry.RetryIf(isRetryableError),
	); retryErr != nil {
		return nil, "", retryErr
	}

	if count > 0 {
		log.Info().Msgf("%s %s succeeded after %d retries", method, url, count)
	}

	return resp, responseBody, nil
}

func isRetryableError(err error) bool {
	msg := err.Error()
	for _, matcher := range unretryableErrors {
		if matcher.MatchString(msg) {
			return false
		}
	}
	return true
}

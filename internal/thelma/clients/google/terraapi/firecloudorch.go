package terraapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"time"
)

// BEE seeding occasionally fails, with Orch occasionally encountering connection timeouts, resets, or DNS errors
// while communicating with some of the services it proxies. (The most common culprits are Agora and Thurloe).
//
// While we investigate this issue, we have configured this client to aggressively retry failed requests that
// seem to be network-related every 30 seconds, for up to 20 minutes, with the goal of determining whether
// the issues resolve on their own after a period of time or require manual intervention (say, a restart of orch).

const defaultRetryAttempts = 40
const defaultRetryDelay = 30 * time.Second

var retryableErrors = []*regexp.Regexp{
	regexp.MustCompile(`java\.net\.SocketTimeoutException`),
	regexp.MustCompile(`java\.net\.UnknownHostException`),
	regexp.MustCompile(`akka\.http\.impl\.engine\.client\.OutgoingConnectionBlueprint\$UnexpectedConnectionClosureException`),
	regexp.MustCompile(`(?m)503 Service Temporarily Unavailable.*nginx`),
	regexp.MustCompile(`503 Service Unavailable`),
}

type FirecloudOrchClient interface {
	RegisterProfile(firstName string, lastName string, title string, contactEmail string, institute string, institutionalProgram string, programLocationCity string, programLocationState string, programLocationCountry string, pi string, nonProfitStatus string) (*http.Response, string, error)
	AgoraMakeMethod(data interface{}) (*http.Response, string, error)
	AgoraSetMethodACLs(name string, namespace string, acls interface{}) (*http.Response, string, error)
	AgoraMakeConfig(data interface{}) (*http.Response, string, error)
	AgoraSetConfigACLs(name string, namespace string, acls interface{}) (*http.Response, string, error)
	AgoraSetNamespaceACLs(namespace string, acls interface{}) (*http.Response, string, error)
}

type firecloudOrchClient struct {
	*terraClient
	appRelease    terra.AppRelease
	retryAttempts uint
	retryDelay    time.Duration
}

func (c *firecloudOrchClient) RegisterProfile(firstName string, lastName string, title string, contactEmail string, institute string, institutionalProgram string, programLocationCity string, programLocationState string, programLocationCountry string, pi string, nonProfitStatus string) (*http.Response, string, error) {
	bodyStruct := struct {
		FirstName              string `json:"firstName"`
		LastName               string `json:"lastName"`
		Title                  string `json:"title"`
		ContactEmail           string `json:"contactEmail"`
		Institute              string `json:"institute"`
		InstitutionalProgram   string `json:"institutionalProgram"`
		ProgramLocationCity    string `json:"programLocationCity"`
		ProgramLocationState   string `json:"programLocationState"`
		ProgramLocationCountry string `json:"programLocationCountry"`
		Pi                     string `json:"pi"`
		NonProfitStatus        string `json:"nonProfitStatus"`
	}{
		FirstName:              firstName,
		LastName:               lastName,
		Title:                  title,
		ContactEmail:           contactEmail,
		Institute:              institute,
		InstitutionalProgram:   institutionalProgram,
		ProgramLocationCity:    programLocationCity,
		ProgramLocationState:   programLocationState,
		ProgramLocationCountry: programLocationCountry,
		Pi:                     pi,
		NonProfitStatus:        nonProfitStatus,
	}
	return c.doJsonRequestWithRetries(http.MethodPost, fmt.Sprintf("%s/register/profile", c.appRelease.URL()), bodyStruct)
}
func (c *firecloudOrchClient) AgoraMakeMethod(data interface{}) (*http.Response, string, error) {
	return c.doJsonRequestWithRetries(http.MethodPost, fmt.Sprintf("%s/api/methods", c.appRelease.URL()), data)
}

func (c *firecloudOrchClient) AgoraSetMethodACLs(name string, namespace string, acls interface{}) (*http.Response, string, error) {
	return c.doJsonRequestWithRetries(http.MethodPost, fmt.Sprintf("%s/api/methods/%s/%s/1/permissions", c.appRelease.URL(), namespace, name), acls)
}

func (c *firecloudOrchClient) AgoraMakeConfig(data interface{}) (*http.Response, string, error) {
	return c.doJsonRequestWithRetries(http.MethodPost, fmt.Sprintf("%s/api/configurations", c.appRelease.URL()), data)
}

func (c *firecloudOrchClient) AgoraSetConfigACLs(name string, namespace string, acls interface{}) (*http.Response, string, error) {
	return c.doJsonRequestWithRetries(http.MethodPost, fmt.Sprintf("%s/api/configurations/%s/%s/1/permissions", c.appRelease.URL(), namespace, name), acls)
}

func (c *firecloudOrchClient) AgoraSetNamespaceACLs(namespace string, acls interface{}) (*http.Response, string, error) {
	return c.doJsonRequestWithRetries(http.MethodPost, fmt.Sprintf("%s/api/configurations/%s/permissions", c.appRelease.URL(), namespace), acls)
}

func (c *firecloudOrchClient) doJsonRequestWithRetries(method string, url string, bodyData interface{}) (*http.Response, string, error) {
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
		return nil, "", fmt.Errorf("error marshalling request body for %s %s: %v", method, url, err)
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
	for _, matcher := range retryableErrors {
		if matcher.MatchString(msg) {
			return true
		}
	}
	return false
}

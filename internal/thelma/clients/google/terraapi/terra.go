package terraapi

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"
	"io"
	"net/http"
)

type TerraClient interface {
	FirecloudOrch(release terra.AppRelease) FirecloudOrchClient
	Sam(release terra.AppRelease) SamClient
	GoogleUserInfo() googleoauth.Userinfo
}

type terraClient struct {
	tokenSource oauth2.TokenSource
	userInfo    googleoauth.Userinfo
	httpClient  http.Client
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

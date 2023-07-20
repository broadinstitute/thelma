package terraapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type SamClient interface {
	FcServiceAccounts([]string, string) (*http.Response, string, error)
	AcceptToS() (*http.Response, string, error)
	UnregisterUser(id string) (*http.Response, string, error)
	CreateCloudExtension(cloud string) (*http.Response, string, error)
}

type samClient struct {
	*terraClient
	appRelease terra.AppRelease
}

func (c *samClient) FcServiceAccounts(memberEmails []string, cloud string) (*http.Response, string, error) {
	bodyStruct := struct {
		MemberEmails []string `json:"memberEmails"`
		Actions      []string `json:"actions"`
		Roles        []string `json:"roles"`
	}{
		MemberEmails: memberEmails,
		Actions:      []string{"get_pet_private_key"},
		Roles:        []string{},
	}
	body, err := json.Marshal(bodyStruct)
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPut, fmt.Sprintf("%s/api/resource/cloud-extension/%s/policies/fc-service-accounts", c.appRelease.URL(), cloud), bytes.NewBuffer(body))
}

func (c *samClient) AcceptToS() (*http.Response, string, error) {
	body, err := json.Marshal("app.terra.bio/#terms-of-service")
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/register/user/v1/termsofservice", c.appRelease.URL()), bytes.NewBuffer(body))
}

func (c *samClient) UnregisterUser(id string) (*http.Response, string, error) {
	return c.doJsonRequest(http.MethodDelete, fmt.Sprintf("%s/api/admin/user/%s", c.appRelease.URL(), id), &bytes.Buffer{})
}

func (c *samClient) CreateCloudExtension(cloud string) (*http.Response, string, error) {
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/api/resources/v2/cloud-extension/%s", c.appRelease.URL(), cloud), &bytes.Buffer{})
}

package terraapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"net/http"
)

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
	appRelease terra.AppRelease
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
	body, err := json.Marshal(bodyStruct)
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/register/profile", c.appRelease.URL()), bytes.NewBuffer(body))
}

func (c *firecloudOrchClient) AgoraMakeMethod(data interface{}) (*http.Response, string, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/api/methods", c.appRelease.URL()), bytes.NewBuffer(body))
}

func (c *firecloudOrchClient) AgoraSetMethodACLs(name string, namespace string, acls interface{}) (*http.Response, string, error) {
	body, err := json.Marshal(acls)
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/api/methods/%s/%s/1/permissions", c.appRelease.URL(), namespace, name), bytes.NewBuffer(body))
}

func (c *firecloudOrchClient) AgoraMakeConfig(data interface{}) (*http.Response, string, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/api/configurations", c.appRelease.URL()), bytes.NewBuffer(body))
}

func (c *firecloudOrchClient) AgoraSetConfigACLs(name string, namespace string, acls interface{}) (*http.Response, string, error) {
	body, err := json.Marshal(acls)
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/api/configurations/%s/%s/1/permissions", c.appRelease.URL(), namespace, name), bytes.NewBuffer(body))
}

func (c *firecloudOrchClient) AgoraSetNamespaceACLs(namespace string, acls interface{}) (*http.Response, string, error) {
	body, err := json.Marshal(acls)
	if err != nil {
		return nil, "", err
	}
	return c.doJsonRequest(http.MethodPost, fmt.Sprintf("%s/api/configurations/%s/permissions", c.appRelease.URL(), namespace), bytes.NewBuffer(body))
}

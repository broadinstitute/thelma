package terraapi

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// BEE seeding occasionally fails, with Orch occasionally encountering connection timeouts, resets, or DNS errors
// while communicating with some of the services it proxies. (The most common culprits are Agora and Thurloe).
//
// While we investigate this issue, we have configured this client to aggressively retry failed requests that
// seem to be network-related every 30 seconds, for up to 20 minutes, with the goal of determining whether
// the issues resolve on their own after a period of time or require manual intervention (say, a restart of orch).

const defaultRetryAttempts = 40
const defaultRetryDelay = 30 * time.Second

// unretryableErrors a list of errors from Orch that should NOT be retried
var unretryableErrors = []*regexp.Regexp{
	regexp.MustCompile(`409 Conflict`),
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

package iap

import "github.com/pkg/errors"

type Project int

const (
	DspDevopsSuperProd Project = iota
	DspToolsK8s
)

func ParseProject(s string) (Project, error) {
	switch s {
	case "dsp-devops-super-prod":
		return DspDevopsSuperProd, nil
	case "dsp-tools-k8s":
		return DspToolsK8s, nil
	}
	return -1, errors.Errorf("unknown project: %s", s)
}

// ProjectID returns the project ID for the given project.
// This is the alphanumeric/hyphenated ID, not the one that can contain spaces or the one that is all numbers.
func (p Project) ProjectID() (string, error) {
	switch p {
	case DspDevopsSuperProd:
		return "dsp-devops-super-prod", nil
	case DspToolsK8s:
		return "dsp-tools-k8s", nil
	}
	return "unknown", errors.Errorf("unknown project enum: %d", p)
}

// tokenKey generates the credentials.TokenProvider key for the given project.
// By way of explicit example, for DspDevopsSuperProd, this function returns "dsp-devops-super-prod-iap-oauth-token",
// which means that a DSP_DEVOPS_SUPER_PROD_IAP_OAUTH_TOKEN environment variable would override that token generation
// and so on.
func (p Project) tokenKey() (string, error) {
	projectID, err := p.ProjectID()
	if err != nil {
		return "", err
	} else {
		return projectID + "-iap-oauth-token", nil
	}
}

// oauthCredentials returns the client ID and client secret (in that order) for the given project.
func (p Project) oauthCredentials(cfg iapConfig) (string, string, error) {
	switch p {
	case DspDevopsSuperProd:
		return cfg.Projects.DspDevopsSuperProd.ClientID, cfg.Projects.DspDevopsSuperProd.ClientSecret, nil
	case DspToolsK8s:
		return cfg.Projects.DspToolsK8s.ClientID, cfg.Projects.DspToolsK8s.ClientSecret, nil
	}
	return "", "", errors.Errorf("unknown project enum: %d", p)
}

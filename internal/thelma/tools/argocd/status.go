package argocd

import (
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"gopkg.in/yaml.v3"
)

// TODO enum types w/ custom unmarshaller for health status ("Degraded", "Progressing", etc) and sync status ("Synced", "OutOfSync")

type application struct {
	Status ApplicationStatus
}

type ApplicationStatus struct {
	// Overall health status for the application
	Health struct {
		Status string
	}
	// Overall sync status for the application
	Sync struct {
		Status string
	}
	Resources []Resource
}

type Resource struct {
	Kind      string
	Name      string
	Version   string
	Namespace string
	Status    string
	Health    struct {
		Status  string `yaml:",omitempty"`
		Message string `yaml:",omitempty"`
	} `yaml:",omitempty"`
}

func (a *argocd) AppStatus(appName string) (ApplicationStatus, error) {
	buf := bytes.Buffer{}
	err := a.runCommand([]string{"app", "get", appName, "-o", "yaml"}, func(options *shell.RunOptions) {
		options.Stdout = &buf
	})
	if err != nil {
		return ApplicationStatus{}, err
	}

	var app application
	if err = yaml.Unmarshal(buf.Bytes(), &app); err != nil {
		return ApplicationStatus{}, fmt.Errorf("error unmarshalling argo app %s: %v", appName, err)
	}

	return app.Status, nil
}

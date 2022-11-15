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
	Health struct {
		Status string
	}
	Sync struct {
		Status string
	}
	Resources []Resource
}

type Resource struct {
	Kind    string
	Name    string
	Status  string
	Version string
	Health  struct {
		Status  string
		Message string
	}
}

func (a *argocd) Status(appName string) (ApplicationStatus, error) {
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

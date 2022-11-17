package argocd

import (
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"gopkg.in/yaml.v3"
)

type HealthStatus int

const (
	Unknown HealthStatus = iota
	Progressing
	Suspended
	Healthy
	Degraded
	Missing
)

func (h *HealthStatus) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case "Unknown":
		*h = Unknown
		return nil
	case "Progressing":
		*h = Progressing
		return nil
	case "Suspended":
		*h = Suspended
		return nil
	case "Healthy":
		*h = Healthy
		return nil
	case "Degraded":
		*h = Degraded
		return nil
	case "Missing":
		*h = Missing
		return nil
	}

	return fmt.Errorf("unknown health status: %v", value.Value)
}

func (h HealthStatus) MarshalYAML() (interface{}, error) {
	str := h.String()
	if str == "" {
		return nil, fmt.Errorf("unknown health status: %v", h)
	}
	return str, nil
}

func (h HealthStatus) String() string {
	switch h {
	case Unknown:
		return "Unknown"
	case Progressing:
		return "Progressing"
	case Suspended:
		return "Suspended"
	case Healthy:
		return "Healthy"
	case Degraded:
		return "Degraded"
	case Missing:
		return "Missing"
	}
	return ""
}

type SyncStatus int

const (
	UnknownSyncStatus SyncStatus = iota
	Synced
	OutOfSync
)

func (s *SyncStatus) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case "Unknown":
		*s = UnknownSyncStatus
		return nil
	case "Synced":
		*s = Synced
		return nil
	case "OutOfSync":
		*s = OutOfSync
		return nil
	}

	return fmt.Errorf("unknown sync status: %v", value.Value)
}

func (s SyncStatus) MarshalYAML() (interface{}, error) {
	str := s.String()
	if str == "" {
		return nil, fmt.Errorf("unknown sync status: %v", s)
	}
	return str, nil
}

func (s SyncStatus) String() string {
	switch s {
	case UnknownSyncStatus:
		return "Unknown"
	case Synced:
		return "Synced"
	case OutOfSync:
		return "OutOfSync"
	}
	return ""
}

type application struct {
	Status ApplicationStatus
}

type ApplicationStatus struct {
	// Overall health status for the application
	Health struct {
		Status HealthStatus
	}
	// Overall sync status for the application
	Sync struct {
		Status SyncStatus
	}
	Resources []Resource
}

type Resource struct {
	Kind      string
	Name      string
	Version   string
	Namespace string
	Status    SyncStatus
	Health    struct {
		Status  HealthStatus `yaml:",omitempty"`
		Message string       `yaml:",omitempty"`
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

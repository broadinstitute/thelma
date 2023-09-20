package argocd

import (
	"github.com/pkg/errors"
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

	return errors.Errorf("unknown health status: %v", value.Value)
}

func (h HealthStatus) MarshalYAML() (interface{}, error) {
	str := h.String()
	if str == "" {
		return nil, errors.Errorf("unknown health status: %v", h)
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

	return errors.Errorf("unknown sync status: %v", value.Value)
}

func (s SyncStatus) MarshalYAML() (interface{}, error) {
	str := s.String()
	if str == "" {
		return nil, errors.Errorf("unknown sync status: %v", s)
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

type Resource struct {
	Kind      string
	Name      string
	Group     string
	Version   string
	Namespace string
	Status    SyncStatus
	Health    *struct {
		Status  HealthStatus `yaml:",omitempty"`
		Message string       `yaml:",omitempty"`
	} `yaml:",omitempty"`
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

type ApplicationSpec struct {
	Source struct {
		TargetRevision string `yaml:"targetRevision"`
	}
}

type application struct {
	Spec   ApplicationSpec
	Status ApplicationStatus
}

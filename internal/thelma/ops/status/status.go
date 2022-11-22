package status

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"time"
)

type Status struct {
	Health             *argocd.HealthStatus `yaml:",omitempty"`
	Sync               *argocd.SyncStatus   `yaml:",omitempty"`
	Error              string               `yaml:",omitempty"`
	UnhealthyResources []Resource           `yaml:"resources,omitempty"`
}

func (s Status) IsHealthy() bool {
	if s.Health == nil {
		return false
	}
	return *s.Health == argocd.Healthy
}

type Resource struct {
	argocd.Resource
	Events []Event `yaml:",omitempty"`
}

type Event struct {
	Count          int32
	Message        string
	Node           string
	FirstTimestamp time.Time
	LastTimestamp  time.Time
	Type           string
}

func (s *Status) Headline() string {
	if s.Error != "" {
		return fmt.Sprintf("error: %v", s.Error)
	}
	if s.Health == nil {
		return "Unknown"
	}

	if len(s.UnhealthyResources) == 0 {
		return s.Health.String()
	}

	// we have some unhealthy resources, pick one that has some events to display in more detail
	resource := s.UnhealthyResources[0]

	if len(resource.Events) == 0 {
		return fmt.Sprintf("%s: %s: %s", s.Health.String(), resource.Name, resource.Health.Message)
	}

	mostRecentEvent := resource.Events[0]
	for _, e := range resource.Events {
		if e.LastTimestamp.After(mostRecentEvent.LastTimestamp) {
			mostRecentEvent = e
		}
	}
	return fmt.Sprintf("%s: %s: %s", s.Health.String(), resource.Name, mostRecentEvent.Message)
}

package status

import (
	"context"
	"fmt"
	k8sclients "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	argocdnames "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sync"
	"time"
)

type Status struct {
	Health             *argocd.HealthStatus
	Sync               *argocd.SyncStatus
	Error              error      `yaml:",omitempty"`
	UnhealthyResources []Resource `yaml:"resources,omitempty"`
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

type Reporter interface {
	Status(release terra.Release) (Status, error)
	Statuses(releases []terra.Release) (map[terra.Release]Status, error)
}

func NewReporter(argocd argocd.ArgoCD, kubeclients k8sclients.Clients) Reporter {
	return &reporter{
		argocd:      argocd,
		kubeclients: kubeclients,
	}
}

type reporter struct {
	argocd      argocd.ArgoCD
	kubeclients k8sclients.Clients
}

func (r *reporter) Status(release terra.Release) (Status, error) {
	appStatus, err := r.argocd.AppStatus(argocdnames.ApplicationName(release))
	if err != nil {
		return errStatus(err)
	}

	healthy := appStatus.Health.Status == argocd.Healthy

	var unhealthyResources []Resource

	if !healthy {
		apiClient, err := r.kubeclients.ForRelease(release)
		if err != nil {
			return errStatus(err)
		}
		eventList, err := apiClient.CoreV1().Events(release.Namespace()).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return errStatus(err)
		}

		for _, argoResource := range appStatus.Resources {
			if argoResource.Health.Status != argocd.Healthy {
				resource := Resource{Resource: argoResource}
				if argoResource.Health.Status != argocd.Missing {
					var podSelector *metav1.LabelSelector
					var resourceUID types.UID

					if argoResource.Kind == "Deployment" {
						deployment, err := apiClient.AppsV1().Deployments(release.Namespace()).Get(context.Background(), argoResource.Name, metav1.GetOptions{})
						if err != nil {
							return errStatus(err)
						}
						podSelector = deployment.Spec.Selector
						resourceUID = deployment.UID
					} else if argoResource.Kind == "Statefulset" {
						sts, err := apiClient.AppsV1().StatefulSets(release.Namespace()).Get(context.Background(), argoResource.Namespace, metav1.GetOptions{})
						if err != nil {
							return errStatus(err)
						}
						podSelector = sts.Spec.Selector
						resourceUID = sts.UID
					} else {
						// not sure how to select pods in whatever type of this resource this is, so don't try
						continue
					}

					selectorMap, err := metav1.LabelSelectorAsMap(podSelector)
					if err != nil {
						return errStatus(err)
					}

					selectorString := labels.SelectorFromSet(selectorMap).String()

					podList, err := apiClient.CoreV1().Pods(release.Namespace()).List(context.Background(), metav1.ListOptions{
						LabelSelector: selectorString,
					})
					if err != nil {
						return Status{}, err
					}
					log.Debug().Msgf("Found %d pods in %s %s", len(podList.Items), argoResource.Kind, argoResource.Name)

					var matchingEvents []corev1.Event
					matchingEvents = append(matchingEvents, r.eventsMatchingUID(eventList.Items, resourceUID)...)
					matchingEvents = append(matchingEvents, r.eventsForPods(eventList.Items, podList.Items)...)

					var events []Event
					for _, k8sEvent := range matchingEvents {
						events = append(events, toEventView(k8sEvent))
					}
					resource.Events = events
				}
				unhealthyResources = append(unhealthyResources, resource)
			}
		}
	}

	return Status{
		Health:             &appStatus.Health.Status,
		Sync:               &appStatus.Sync.Status,
		UnhealthyResources: unhealthyResources,
	}, nil
}

func (s *Status) Headline() string {
	if s.Error != nil {
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

func (r *reporter) Statuses(releases []terra.Release) (map[terra.Release]Status, error) {
	statuses := make(map[terra.Release]Status)
	var mutex sync.Mutex

	var jobs []pool.Job

	for _, unsafe := range releases {
		release := unsafe // copy invariant to tmp variable

		jobs = append(jobs, pool.Job{
			Name: release.FullName(),
			Run: func(_ pool.StatusReporter) error {
				status, err := r.Status(release)
				if err != nil {
					return fmt.Errorf("error generating status report for %s: %v", release.Name(), err)
				}
				mutex.Lock()
				statuses[release] = status
				mutex.Unlock()
				return nil
			},
		})
	}

	_pool := pool.New(jobs, func(options *pool.Options) {
		options.NumWorkers = 10
		options.Summarizer.WorkDescription = "services checked"
	})
	err := _pool.Execute()
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (r *reporter) eventsForPods(events []corev1.Event, pods []corev1.Pod) []corev1.Event {
	var podEvents []corev1.Event
	for _, pod := range pods {
		matches := r.eventsMatchingUID(events, pod.UID)
		podEvents = append(podEvents, matches...)
	}
	return podEvents
}

func (r *reporter) eventsMatchingUID(events []corev1.Event, uid types.UID) []corev1.Event {
	var matches []corev1.Event

	for _, event := range events {
		if event.InvolvedObject.UID == uid {
			matches = append(matches, event)
		}
	}
	return matches
}

func toEventView(event corev1.Event) Event {
	return Event{
		Count:          event.Count,
		Message:        event.Message,
		Node:           event.Source.Host,
		FirstTimestamp: event.FirstTimestamp.Time,
		LastTimestamp:  event.LastTimestamp.Time,
		Type:           event.Type,
	}
}

// indicates an error was encountered while retrieving the status
func errStatus(err error) (Status, error) {
	return Status{
		Error: err,
	}, err
}

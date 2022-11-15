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
	Healthy            bool
	Synced             bool
	UnhealthyResources []Resource `yaml:",omitempty"`
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
		return Status{}, err
	}

	synced := appStatus.Sync.Status == "Synced"
	healthy := appStatus.Health.Status == "Healthy"

	var unhealthyResources []Resource

	if !healthy {
		apiClient, err := r.kubeclients.ForRelease(release)
		if err != nil {
			return Status{}, err
		}
		eventList, err := apiClient.CoreV1().Events(release.Namespace()).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return Status{}, err
		}

		for _, argoResource := range appStatus.Resources {
			if argoResource.Health.Status != "" && argoResource.Health.Status != "Healthy" {
				resource := Resource{Resource: argoResource}

				if argoResource.Health.Status != "Missing" {
					var podSelector *metav1.LabelSelector
					var resourceUID types.UID

					if argoResource.Kind == "Deployment" {
						deployment, err := apiClient.AppsV1().Deployments(release.Namespace()).Get(context.Background(), argoResource.Name, metav1.GetOptions{})
						if err != nil {
							return Status{}, err
						}
						podSelector = deployment.Spec.Selector
						resourceUID = deployment.UID
					} else if argoResource.Kind == "Statefulset" {
						sts, err := apiClient.AppsV1().StatefulSets(release.Namespace()).Get(context.Background(), argoResource.Namespace, metav1.GetOptions{})
						if err != nil {
							return Status{}, err
						}
						podSelector = sts.Spec.Selector
						resourceUID = sts.UID
					} else {
						// not sure how to select pods in whatever type of this resource this is, so don't try
						continue
					}

					selectorMap, err := metav1.LabelSelectorAsMap(podSelector)
					if err != nil {
						return Status{}, err
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
		Healthy:            healthy,
		Synced:             synced,
		UnhealthyResources: unhealthyResources,
	}, nil
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

func (r *reporter) Statuses(releases []terra.Release) (map[terra.Release]Status, error) {
	statuses := make(map[terra.Release]Status)
	var mutex sync.Mutex

	var jobs []pool.Job

	for _, unsafe := range releases {
		release := unsafe // copy invariant to tmp variable

		jobs = append(jobs, pool.Job{
			Name: argocdnames.ApplicationName(release),
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
	})
	err := _pool.Execute()
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

package status

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type eventMatcher struct {
	events    []corev1.Event
	namespace string
	apiClient kubernetes.Interface
}

func newEventMatcher(apiClient kubernetes.Interface, namespace string) (*eventMatcher, error) {
	events, err := apiClient.CoreV1().Events(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return &eventMatcher{
		events:    events.Items,
		namespace: namespace,
		apiClient: apiClient,
	}, nil
}

// annotateResource with any matching events
func (e *eventMatcher) annotateResourceWithMatchingEvents(resource *Resource) error {
	if resource.Namespace != e.namespace {
		log.Warn().Msgf("Unexpected data inconsistency: resource namespace %s does not match event cache namespace %s", resource.Namespace, e.namespace)
		return nil
	}

	var podSelector *metav1.LabelSelector
	var resourceUID types.UID

	if resource.Kind == "Deployment" {
		deployment, err := e.apiClient.AppsV1().Deployments(resource.Namespace).Get(context.Background(), resource.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		podSelector = deployment.Spec.Selector
		resourceUID = deployment.UID
	} else if resource.Kind == "Statefulset" {
		sts, err := e.apiClient.AppsV1().StatefulSets(resource.Namespace).Get(context.Background(), resource.Namespace, metav1.GetOptions{})
		if err != nil {
			return err
		}
		podSelector = sts.Spec.Selector
		resourceUID = sts.UID
	} else {
		// not sure how to select pods in whatever type of this resource this is, so don't try
		return nil
	}

	selectorMap, err := metav1.LabelSelectorAsMap(podSelector)
	if err != nil {
		return err
	}

	selectorString := labels.SelectorFromSet(selectorMap).String()

	podList, err := e.apiClient.CoreV1().Pods(resource.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selectorString,
	})
	if err != nil {
		return err
	}
	log.Debug().Msgf("Found %d pods in %s %s", len(podList.Items), resource.Kind, resource.Name)

	var matchingEvents []corev1.Event
	matchingEvents = append(matchingEvents, e.eventsMatchingUID(resourceUID)...)
	matchingEvents = append(matchingEvents, e.eventsForPods(podList.Items)...)

	var events []Event
	for _, k8sEvent := range matchingEvents {
		events = append(events, toEventView(k8sEvent))
	}
	resource.Events = events
	return nil
}

func (e *eventMatcher) eventsForPods(pods []corev1.Pod) []corev1.Event {
	var podEvents []corev1.Event
	for _, pod := range pods {
		matches := e.eventsMatchingUID(pod.UID)
		podEvents = append(podEvents, matches...)
	}
	return podEvents
}

func (e *eventMatcher) eventsMatchingUID(uid types.UID) []corev1.Event {
	var matches []corev1.Event

	for _, event := range e.events {
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

package status

import (
	k8sclients "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	argocdnames "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"sync"
)

// Reader can read the status of a release from Kubernetes and ArgoCD
type Reader interface {
	Status(release terra.Release) (*Status, error)
	Statuses(releases []terra.Release) (map[terra.Release]*Status, error)
}

func NewReporter(argocd argocd.ArgoCD, kubeclients k8sclients.Clients) Reader {
	return &reader{
		argocd:      argocd,
		kubeclients: kubeclients,
	}
}

type reader struct {
	argocd      argocd.ArgoCD
	kubeclients k8sclients.Clients
}

func (r *reader) Status(release terra.Release) (*Status, error) {
	appStatus, err := r.argocd.AppStatus(argocdnames.ApplicationName(release))
	if err != nil {
		return nil, err
	}

	status := Status{
		Health:             appStatus.Health.Status,
		Sync:               appStatus.Sync.Status,
		UnhealthyResources: r.buildUnhealthyResourceList(appStatus, release),
	}

	return &status, nil
}

func (r *reader) Statuses(releases []terra.Release) (map[terra.Release]*Status, error) {
	statuses := make(map[terra.Release]*Status)
	var mutex sync.Mutex

	var jobs []pool.Job

	for _, unsafe := range releases {
		release := unsafe // copy invariant to tmp variable

		jobs = append(jobs, pool.Job{
			Name: release.FullName(),
			Run: func(_ pool.StatusReporter) error {
				status, err := r.Status(release)
				if err != nil {
					return errors.Errorf("error generating status report for %s: %v", release.Name(), err)
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
		options.LogSummarizer.WorkDescription = "services checked"
	})
	err := _pool.Execute()
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (r *reader) buildUnhealthyResourceList(appStatus argocd.ApplicationStatus, release terra.Release) []Resource {
	var unhealthyResources []Resource

	for _, argoResource := range appStatus.Resources {
		if argoResource.Health == nil || argoResource.Health.Status == argocd.Healthy {
			continue
		}
		unhealthyResources = append(unhealthyResources, Resource{Resource: argoResource})
	}

	_eventMatcher, err := r.buildEventMatcher(release)
	if err != nil {
		log.Debug().Err(err).Msgf("failed to load events from Kubernetes API for %s; disabling rich status reports", release.FullName())
		return unhealthyResources
	}

	for i, resource := range unhealthyResources {
		if err := _eventMatcher.annotateResourceWithMatchingEvents(&resource); err != nil {
			log.Warn().Err(err).Msgf("failed to load events for %s from Kubernetes API in %s", resource.Name, release.FullName())
			continue
		}
		unhealthyResources[i] = resource
	}

	return unhealthyResources
}

func (r *reader) buildEventMatcher(release terra.Release) (*eventMatcher, error) {
	apiClient, err := r.kubeclients.ForRelease(release)
	if err != nil {
		return nil, err
	}
	return newEventMatcher(apiClient, release.Namespace())
}

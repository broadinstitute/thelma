package sync

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/status"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	naming "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/rs/zerolog/log"
	"sync"
)

type Sync interface {
	// Sync will sync the Argo app(s) for a set of releases, wait for them to be healthy,
	// and generate and return status report.
	Sync(releases []terra.Release, maxParallel int, options ...argocd.SyncOption) (map[terra.Release]status.Status, error)
}

func New(argocd argocd.ArgoCD) Sync {
	return &syncer{argocd: argocd}
}

type syncer struct {
	argocd argocd.ArgoCD
	status status.Reporter
}

func extractWaitHealthy(opts []argocd.SyncOption) bool {
	var options argocd.SyncOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options.WaitHealthy
}

func (s *syncer) Sync(releases []terra.Release, maxParallel int, options ...argocd.SyncOption) (map[terra.Release]status.Status, error) {
	var jobs []pool.Job

	waitHealthy := extractWaitHealthy(options)

	optionsNoWaitHealthy := append(options, func(options *argocd.SyncOptions) {
		options.WaitHealthy = false
	})

	var statusMap sync.Map

	for _, release := range releases {
		r := release
		jobs = append(jobs, pool.Job{
			Name: naming.ApplicationName(r),
			Run: func(_ pool.StatusReporter) error {
				log.Info().Msgf("Syncing ArgoCD application(s) for %s in %s", r.Name(), r.Destination().Name())
				if err := s.argocd.SyncRelease(r, optionsNoWaitHealthy...); err != nil {
					return err
				}

				var waitErr error
				if waitHealthy {
					waitErr = s.argocd.WaitHealthy(r)
				}
				_status, statusErr := s.status.Status(r)

				if statusErr != nil {
					if waitErr != nil {
						log.Err(statusErr).Msgf("error computing status for %s", r.FullName())
					} else {
						return statusErr
					}
				}

				statusMap.Store(r, _status)
				return waitErr
			},
		})
	}

	_pool := pool.New(jobs, func(options *pool.Options) {
		options.NumWorkers = maxParallel
		options.StopProcessingOnError = false
	})

	err := _pool.Execute()

	result := make(map[terra.Release]status.Status)
	statusMap.Range(func(key, value any) bool {
		release, ok := key.(terra.Release)
		if !ok {
			panic("type assertion failed")
		}
		_status, ok := value.(status.Status)
		if !ok {
			panic("type assertion failed")
		}
		result[release] = _status
		return true
	})

	return result, err
}

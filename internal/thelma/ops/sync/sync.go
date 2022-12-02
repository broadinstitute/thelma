package sync

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/status"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/rs/zerolog/log"
	"sort"
	"sync"
	"time"
)

const waitHealthyPollingInterval = 30 * time.Second

type Sync interface {
	// Sync will sync the Argo app(s) for a set of releases, wait for them to be healthy,
	// and generate and return status reports (useful for understanding why a sync failed).
	Sync(releases []terra.Release, maxParallel int, options ...argocd.SyncOption) (map[terra.Release]*status.Status, error)
}

func New(argocd argocd.ArgoCD, status status.Reporter) Sync {
	return &syncer{
		argocd: argocd,
		status: status,
	}
}

type syncer struct {
	argocd argocd.ArgoCD
	status status.Reporter
}

// Sync a set of releases and return a status report indicating whether the release is healthy.
func (s *syncer) Sync(releases []terra.Release, maxParallel int, options ...argocd.SyncOption) (map[terra.Release]*status.Status, error) {
	var jobs []pool.Job

	waitHealthyTimeout := s.extractWaitHealthy(options)

	optionsNoWaitHealthy := withOption(options, func(options *argocd.SyncOptions) {
		options.WaitHealthy = false
	})

	destination, hasSingleDestination := checkIfSingleDestination(releases)

	statusMap := make(map[terra.Release]*status.Status)
	var mutex sync.Mutex

	for _, unsafe := range releases {
		release := unsafe

		jobName := release.FullName()
		if hasSingleDestination {
			jobName = release.Name()
		}

		jobs = append(jobs, pool.Job{
			Name: jobName,
			Run: func(statusReporter pool.StatusReporter) error {
				opts := withOption(optionsNoWaitHealthy, func(options *argocd.SyncOptions) {
					options.StatusReporter = statusReporter
				})
				if err := s.argocd.SyncRelease(release, opts...); err != nil {
					return err
				}

				_status, err := s.waitHealthy(release, waitHealthyTimeout, statusReporter)

				mutex.Lock()
				statusMap[release] = _status
				mutex.Unlock()

				return err
			},
		})
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Name < jobs[j].Name
	})

	_pool := pool.New(jobs, func(options *pool.Options) {
		options.NumWorkers = maxParallel
		options.StopProcessingOnError = false
		options.Summarizer.WorkDescription = "services synced"

		if hasSingleDestination {
			options.Summarizer.Footer = fmt.Sprintf("Check status in ArgoCD at %s", s.argocd.DestinationURL(destination))
		}

		options.Metrics.Enabled = true
		options.Metrics.PoolName = "ops_sync"
	})

	err := _pool.Execute()

	return statusMap, err
}

// waitHealthy waits for a release's primary ArgoCD application to be healthy. If:
// an unknown error is encountered while generating the status report (because, say, ArgoCD is down):
// .     we return nil + the underlying error
// the application becomes healthy within the timeout:
// .     we return the status report and nil error
// the application does not become healthy within the timeout:
// .     we return the status report + a timeout error
func (s *syncer) waitHealthy(release terra.Release, maxWait time.Duration, statusReporter pool.StatusReporter) (*status.Status, error) {
	lastStatus, err := s.status.Status(release)
	if err != nil {
		// we failed to retrieve status, so return an error
		return nil, err
	}
	updateStatus(lastStatus, statusReporter)
	if lastStatus.IsHealthy() {
		return lastStatus, nil
	}

	if maxWait == 0 {
		log.Debug().Msgf("Not waiting for %s to be healthy", release.FullName())
		return lastStatus, nil
	}

	for {
		timeout := time.After(maxWait)
		ticker := time.NewTicker(waitHealthyPollingInterval)

		for {
			select {
			case <-timeout:
				return lastStatus, fmt.Errorf("timed out waiting for healthy (%s)", lastStatus.Headline())
			case <-ticker.C:
				lastStatus, err = s.status.Status(release)
				if err != nil {
					return nil, err
				}
				updateStatus(lastStatus, statusReporter)
				if lastStatus.IsHealthy() {
					return lastStatus, nil
				}
			}
		}
	}
}

func updateStatus(status *status.Status, statusReporter pool.StatusReporter) {
	statusReporter.Update(pool.Status{Message: status.Headline()})
}

func checkIfSingleDestination(releases []terra.Release) (terra.Destination, bool) {
	var destination terra.Destination
	for _, release := range releases {
		if destination == nil {
			destination = release.Destination()
		}
		if release.Destination().Name() != destination.Name() {
			return nil, false
		}
	}
	return destination, true
}

func (s *syncer) extractWaitHealthy(opts []argocd.SyncOption) time.Duration {
	options := s.argocd.DefaultSyncOptions()
	for _, opt := range opts {
		opt(&options)
	}
	if !options.WaitHealthy {
		return 0
	}
	return time.Duration(options.WaitHealthyTimeoutSeconds) * time.Second
}

func withOption(opts []argocd.SyncOption, option ...argocd.SyncOption) []argocd.SyncOption {
	var result []argocd.SyncOption
	result = append(result, opts...)
	result = append(result, option...)
	return result
}

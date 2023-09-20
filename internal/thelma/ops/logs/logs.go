package logs

import (
	"fmt"
	k8sclients "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

const defaultNumWorkers = 10

// Logs exports container logs from a Kubernetes cluster
type Logs interface {
	Logs(release terra.Release, option ...LogsOption) error
	Export(releases []terra.Release, option ...ExportOption) (map[terra.Release]artifacts.Location, error)
}

// container struct representing a container definition for which logs should be collected
type container struct {
	// containerName name of the container. Eg. "rawls-backend-sqlproxy"
	containerName string
	// resourceKind name of the resource type the container belongs to. Eg. "deployment", "statefulset"
	resourceKind string
	// resourceName name of the resource the container belongs to. Eg. "rawls-backend-deployment"
	resourceName string
	// release reference to the terra.Release that this container belongs to
	release terra.Release
	// podSelectorLabels labels the container's replicaset uses to manage its containers
	podSelectorLabels map[string]string
	// kubectx associated with this container
	kubectx kubecfg.Kubectx
}

type ExportOption func(options *ExportOptions)

type ExportOptions struct {
	// ParallelWorkers Number of goroutines to spawn to export container logs in parallel
	ParallelWorkers int
	// Artifacts options governing artifact storage
	Artifacts artifacts.Options
	LogsOptions
}

type LogsOption func(optins *LogsOptions)

type LogsOptions struct {
	// ContainerFilter optional function for filtering which container logs are exported
	ContainerFilter func(container) bool
	// MaxLines maximum number of log lines to retrieve
	MaxLines int
}

func New(k8sclients k8sclients.Clients, artifacts artifacts.Artifacts) Logs {
	return &logs{
		k8sclients: k8sclients,
		artifacts:  artifacts,
	}
}

type logs struct {
	k8sclients k8sclients.Clients
	artifacts  artifacts.Artifacts
}

func (l *logs) Logs(release terra.Release, opts ...LogsOption) error {
	options := defaultLogsOptions()
	for _, opt := range opts {
		opt(&options)
	}

	containers, err := l.buildContainerList(release)
	if err != nil {
		return err
	}

	containers = filterContainers(containers, options.ContainerFilter)

	if len(containers) == 0 {
		return errors.Errorf("found no matching containers for %s in %s", release.Name(), release.Destination().Name())
	}
	var _container container
	if len(containers) == 1 {
		_container = containers[0]
	} else {
		_container, err = tryToPickCorrectContainer(containers)
		if err != nil {
			return err
		}
		log.Info().Msgf("Found %d matching containers in %s, will collect logs for %s in %s", len(containers), release.Name(), _container.containerName, _container.resourceName)
	}

	_kubectl, err := l.k8sclients.Kubectl()
	if err != nil {
		return err
	}

	return _kubectl.Logs(_container.kubectx, _container.podSelectorLabels, func(opts *kubectl.LogsOptions) {
		opts.ContainerName = _container.containerName
		opts.MaxLines = options.MaxLines
	})
}

func (l *logs) Export(releases []terra.Release, opts ...ExportOption) (map[terra.Release]artifacts.Location, error) {
	options := ExportOptions{
		ParallelWorkers: defaultNumWorkers,
		LogsOptions:     defaultLogsOptions(),
	}
	for _, opt := range opts {
		opt(&options)
	}

	artifactMgr := l.artifacts.NewManager(artifacts.ContainerLog, options.Artifacts)
	nameGenerator := newLogNameGenerator()

	_kubectl, err := l.k8sclients.Kubectl()
	if err != nil {
		return nil, err
	}

	containers, err := l.buildContainerList(releases...)
	if err != nil {
		return nil, err
	}

	containers = filterContainers(containers, options.ContainerFilter)

	var jobs []pool.Job
	for _, unsafe := range containers {
		// copy loop variant into local variable for safety
		_container := unsafe

		jobs = append(jobs, pool.Job{
			Name: _container.containerName,
			Run: func(_ pool.StatusReporter) error {
				logName := nameGenerator.generateName(_container)
				writer, err := artifactMgr.Writer(_container.release, logName)
				if err != nil {
					return err
				}

				runErr := _kubectl.Logs(_container.kubectx, _container.podSelectorLabels, func(logopts *kubectl.LogsOptions) {
					logopts.Writer = writer
					logopts.ContainerName = _container.containerName
					logopts.MaxLines = options.MaxLines
				})

				closeErr := writer.Close()

				if closeErr != nil {
					// prioritize returning close error because it's likely what lead to the logs command failing
					if runErr != nil {
						log.Err(runErr).Msgf("error running `kubectl logs`")
					}
					return closeErr
				}
				return runErr
			},
		})
	}

	err = pool.New(jobs, func(opts *pool.Options) {
		opts.NumWorkers = options.ParallelWorkers
		opts.Summarizer.WorkDescription = "container logs exported"
		opts.Summarizer.Interval = 10 * time.Second
	}).Execute()

	locations := make(map[terra.Release]artifacts.Location)
	for _, release := range releases {
		locations[release] = artifactMgr.BaseLocationForRelease(release)
	}

	return locations, err
}

func tryToPickCorrectContainer(containers []container) (container, error) {
	if len(containers) == 1 {
		return containers[0], nil
	}

	deploymentsOnly := func(c container) bool {
		return c.resourceKind == "deployment"
	}

	specificDeployment := func(c container) bool {
		if strings.Contains(c.resourceName, "backend") {
			return true
		}
		if strings.Contains(c.resourceName, "runner") {
			return true
		}
		return false
	}

	namedoesNotIncludeProxy := func(c container) bool {
		return !strings.Contains(c.containerName, "proxy")
	}

	nameIncludesAppOrAPI := func(c container) bool {
		if strings.Contains(c.containerName, "app") {
			return true
		}
		if strings.Contains(c.containerName, "api") {
			return true
		}
		return c.containerName == "app"
	}

	filtersToTry := []func(container) bool{
		deploymentsOnly,
		specificDeployment,
		namedoesNotIncludeProxy,
		nameIncludesAppOrAPI,
	}

	winnowed := make([]container, len(containers))
	copy(winnowed, containers)

	for _, filter := range filtersToTry {
		filtered := filterContainers(winnowed, filter)
		if len(filtered) == 0 {
			// skip this filter, it removed all possible matches
			continue
		}
		if len(filtered) == 1 {
			return filtered[0], nil
		}
		winnowed = filtered
	}

	var list []string
	for _, _container := range containers {
		list = append(list, fmt.Sprintf("%s in %s", _container.containerName, _container.resourceName))
	}
	return container{}, errors.Errorf("found multiple matching containers, please use flags to specify:\n%s", strings.Join(list, "\n"))
}

func defaultLogsOptions() LogsOptions {
	return LogsOptions{
		ContainerFilter: func(container container) bool {
			return true
		},
		MaxLines: -1, // -1 means no default line limit, i.e. try to export everything
	}
}

func filterContainers(toFilter []container, filterFn func(container container) bool) []container {
	var result []container
	for _, container := range toFilter {
		if filterFn(container) {
			result = append(result, container)
		}
	}
	return result
}

func (l *logs) buildContainerList(releases ...terra.Release) ([]container, error) {
	_kubecfg, err := l.k8sclients.Kubecfg()
	if err != nil {
		return nil, err
	}

	cache := newResourceCache(l.k8sclients)

	var containers []container

	for _, release := range releases {
		kubectx, err := _kubecfg.ForRelease(release)
		if err != nil {
			return nil, err
		}

		resources, err := cache.get(kubectx)
		if err != nil {
			return nil, err
		}

		releaseContainers := selectContainersForRelease(resources, release, kubectx)

		containers = append(containers, releaseContainers...)
	}
	return containers, nil
}

func selectContainersForRelease(resources cacheEntry, release terra.Release, kubectx kubecfg.Kubectx) []container {
	releaseSelector := argocd.ApplicationSelector(argocd.ApplicationName(release))

	var containers []container
	for _, deployment := range resources.deployments {
		if !isSuperset(deployment.GetLabels(), releaseSelector) {
			continue
		}

		containerSpecs := deployment.Spec.Template.Spec.Containers
		for _, containerSpec := range containerSpecs {
			containers = append(containers, container{
				containerName:     containerSpec.Name,
				resourceKind:      deployment.Kind,
				resourceName:      deployment.Name,
				release:           release,
				podSelectorLabels: deployment.Spec.Selector.MatchLabels,
				kubectx:           kubectx,
			})
		}
	}

	// TODO this is a duplicate of the above loop, maybe there's a way to accomplish more generically?
	for _, statefulset := range resources.statefulsets {
		if !isSuperset(statefulset.GetLabels(), releaseSelector) {
			continue
		}

		containerSpecs := statefulset.Spec.Template.Spec.Containers
		for _, containerSpec := range containerSpecs {
			containers = append(containers, container{
				containerName:     containerSpec.Name,
				resourceKind:      statefulset.Kind,
				resourceName:      statefulset.Name,
				release:           release,
				podSelectorLabels: statefulset.Spec.Selector.MatchLabels,
				kubectx:           kubectx,
			})
		}
	}

	return containers
}

// returns true if maybeSuper is a superset of maybeSub, meaning it includes all of the kv pairs in maybeSub
func isSuperset(maybeSuper map[string]string, maybeSub map[string]string) bool {
	for requiredKey, requiredValue := range maybeSub {
		value, exists := maybeSuper[requiredKey]
		if !exists {
			return false
		}
		if value != requiredValue {
			return false
		}
	}
	return true
}

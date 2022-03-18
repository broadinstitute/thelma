// Package render contains code for rendering Kubernetes manifests from Helm charts
package render

import (
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
	"github.com/broadinstitute/thelma/internal/thelma/render/resolver"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
	"time"
)

// multiRenderTimeout how long to wait before timing out a multi render
const multiRenderTimeout = 5 * time.Minute

// Options encapsulates CLI options for a render
type Options struct {
	Releases        []terra.Release // Releases list of releases that will be rendered
	ReleaseScoped   bool            // ReleaseScoped true implies we are rendering a specific release, like leonardo, and not all releases in a cluster or env
	Stdout          bool            // Stdout if true, render to stdout instead of output directory
	OutputDir       string          // OutputDir output directory where manifests should be rendered
	ChartSourceDir  string          // ChartSourceDir path on filesystem where chart sources live
	ResolverMode    resolver.Mode   // ResolverMode resolver mode
	ParallelWorkers int             // ParallelWorkers number of parallel workers
}

// multiRender renders manifests for multiple environments and clusters
type multiRender struct {
	options    *Options             // Options global render options
	state      terra.State          // state terra state provider for looking up environments, clusters, and releases
	configRepo *helmfile.ConfigRepo // configRepo reference to use for executing `helmfile template`
}

// renderError represents an error encountered while rendering for a particular destination
type renderError struct {
	description string // description for the render job that generated in this error
	err         error  // error
}

// renderJob represents a single helmfile invocation with parameters
type renderJob struct {
	description string
	callback    func() error
}

// prefix for configuration settings
const configPrefix = "render"

// renderConfig configuration struct for render
type renderConfig struct {
	Helmfile struct {
		LogLevel string `default:"info" validate:"oneof=debug info warn error"`
	}
}

// DoRender constructs a multiRender and invokes all functions in correct order to perform a complete
// render.
func DoRender(app app.ThelmaApp, globalOptions *Options, helmfileArgs *helmfile.Args) error {
	r, err := newRender(app, globalOptions)
	if err != nil {
		return err
	}
	if err = r.configRepo.CleanOutputDirectoryIfEnabled(); err != nil {
		return err
	}
	if err = r.configRepo.HelmUpdate(); err != nil {
		return err
	}
	if err = r.renderAll(helmfileArgs); err != nil {
		return err
	}
	return nil
}

// newRender is a constructor for Render objects
func newRender(app app.ThelmaApp, options *Options) (*multiRender, error) {
	r := new(multiRender)
	r.options = options

	state, err := app.State()
	if err != nil {
		return nil, err
	}
	r.state = state

	chartCacheDir, err := app.Scratch().Mkdir("chart-cache")
	if err != nil {
		return nil, err
	}

	scratchDir, err := app.Scratch().Mkdir("helmfile")
	if err != nil {
		return nil, err
	}

	cfg := &renderConfig{}
	if err = app.Config().Unmarshal(configPrefix, cfg); err != nil {
		return nil, err
	}

	r.configRepo = helmfile.NewConfigRepo(helmfile.Options{
		ThelmaHome:       app.Config().Home(),
		ChartCacheDir:    chartCacheDir,
		ChartSourceDir:   options.ChartSourceDir,
		ResolverMode:     options.ResolverMode,
		HelmfileLogLevel: cfg.Helmfile.LogLevel,
		Stdout:           options.Stdout,
		OutputDir:        options.OutputDir,
		ScratchDir:       scratchDir,
		ShellRunner:      app.ShellRunner(),
	})

	return r, nil
}

// renderAll renders manifests based on supplied arguments
func (r *multiRender) renderAll(helmfileArgs *helmfile.Args) error {
	jobs, err := r.getJobs(helmfileArgs)
	if err != nil {
		return err
	}
	if len(jobs) == 0 {
		return fmt.Errorf("no matching releases found")
	}

	numWorkers := 1
	if r.options.ParallelWorkers >= 1 {
		numWorkers = r.options.ParallelWorkers
	}
	if len(jobs) < numWorkers {
		// don't make more workers than we have items to process
		numWorkers = len(jobs)
	}

	log.Info().Msgf("Rendering %d release(s) with %d worker(s)", len(jobs), numWorkers)

	queueCh := make(chan renderJob, len(jobs))
	for _, job := range jobs {
		queueCh <- job
	}
	close(queueCh)

	errCh := make(chan renderError, len(jobs))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		id := i
		wg.Add(1)

		logger := log.With().Str("wid", fmt.Sprintf("worker-%d", id)).Logger()

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					logger.Debug().Msg("short circuit triggered, returning")
					return

				case job, ok := <-queueCh:
					if !ok {
						logger.Debug().Msg("no jobs left, returning")
						return
					}

					logger.Debug().Msgf("rendering %s", job.description)
					if err := job.callback(); err != nil {
						logger.Error().Msgf("error rendering %s:\n%v", job.description, err)
						errCh <- renderError{
							description: job.description,
							err:         err,
						}
						cancel()
						return
					}
				}
			}
		}()
	}

	// Wait for workers to finish in a separate goroutine so that we can implement
	// a timeout
	waitCh := make(chan struct{})
	go func() {
		logger := log.With().Str("wid", "wait").Logger()
		logger.Debug().Msg("Waiting for render workers to finish")
		wg.Wait()
		logger.Debug().Msgf("Render workers finished")
		close(errCh)
		close(waitCh)
	}()

	// Block until the wait group is done or we time out.
	logger := log.With().Str("wid", "main").Logger()

	select {
	case <-waitCh:
		logger.Debug().Msgf("multi-render finished")
	case <-time.After(multiRenderTimeout):
		err := fmt.Errorf("[main] multi-render timed out after %s", multiRenderTimeout)
		logger.Error().Err(err)
		return err
	}

	return readErrorsFromChannel(errCh)
}

// getJobs returns the set of render jobs that should be run
func (r *multiRender) getJobs(helmfileArgs *helmfile.Args) ([]renderJob, error) {
	var jobs []renderJob

	for _, release := range r.options.Releases {
		_r := release
		jobs = append(jobs, renderJob{
			description: fmt.Sprintf("release %s in %s %s", _r.Name(), _r.Destination().Type(), _r.Destination().Name()),
			callback: func() error {
				return r.configRepo.RenderForRelease(_r, helmfileArgs)
			},
		})
	}

	if !r.options.ReleaseScoped && helmfileArgs.ArgocdMode {
		// build set of unique destinations from our collection of releases
		destinations := make(map[string]terra.Destination)
		for _, release := range r.options.Releases {
			destination := release.Destination()
			key := fmt.Sprintf("%s:%s", destination.Type().String(), destination.Name())
			if _, exists := destinations[key]; !exists {
				destinations[key] = destination
			}
		}

		// for each unique destination, make a job to render global resources for it
		for _, _destination := range destinations {
			d := _destination
			jobs = append(jobs, renderJob{
				description: fmt.Sprintf("global resources for %s %s", d.Type(), d.Name()),
				callback: func() error {
					return r.configRepo.RenderForDestination(d, helmfileArgs)
				},
			})
		}
	}

	return jobs, nil
}

// readErrorsFromChannel aggregates all errors into a single mega-error
func readErrorsFromChannel(errCh <-chan renderError) error {
	var count int
	var sb strings.Builder

	for {
		renderErr, ok := <-errCh
		if !ok {
			// channel closed, no more results to read
			break
		}
		count++
		description := renderErr.description
		err := renderErr.err
		sb.WriteString(fmt.Sprintf("%s: %v\n", description, err))
	}

	if count > 0 {
		return fmt.Errorf("%d render errors:\n%s", count, sb.String())
	}

	return nil
}

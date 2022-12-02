// Package render contains code for rendering Kubernetes manifests from Helm charts
package render

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
	"github.com/broadinstitute/thelma/internal/thelma/render/resolver"
	"github.com/broadinstitute/thelma/internal/thelma/render/scope"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/rs/zerolog/log"
)

// Options encapsulates CLI options for a render
type Options struct {
	Releases        []terra.Release // Releases list of releases that will be rendered
	Scope           scope.Scope     // Scope indicates whether to render release-specific resources, destination-specific resources, or both
	Stdout          bool            // Stdout if true, render to stdout instead of output directory
	OutputDir       string          // OutputDir output directory where manifests should be rendered
	DebugMode       bool            // DebugMode if true, pass --debug to helmfile to render out invalid manifests
	ChartSourceDir  string          // ChartSourceDir path on filesystem where chart sources live
	ResolverMode    resolver.Mode   // ResolverMode resolver mode
	ParallelWorkers int             // ParallelWorkers number of parallel workers
}

// multiRender renders manifests for multiple environments and clusters
type multiRender struct {
	options    *Options             // Options global render options
	state      terra.State          // state terra state provider for looking up environments, clusters, and releases
	metrics    metrics.Metrics      // metrics client for recording job metrics
	configRepo *helmfile.ConfigRepo // configRepo reference to use for executing `helmfile template`
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

	r.metrics = app.Metrics()

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
		DebugMode:        options.DebugMode,
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

	_pool := pool.New(jobs, func(options *pool.Options) {
		options.Summarizer.Enabled = false

		if r.options.ParallelWorkers >= 1 {
			options.NumWorkers = r.options.ParallelWorkers
		}

		options.Metrics.Enabled = true
		options.Metrics.Client = r.metrics
		options.Metrics.PoolName = "render"
	})

	log.Info().Msgf("Rendering %d release(s) with %d worker(s)", len(jobs), _pool.NumWorkers())
	return _pool.Execute()
}

// getJobs returns the set of render jobs that should be run
func (r *multiRender) getJobs(helmfileArgs *helmfile.Args) ([]pool.Job, error) {
	var jobs []pool.Job

	if r.options.Scope != scope.Destination {
		for _, release := range r.options.Releases {
			_r := release
			jobs = append(jobs, pool.Job{
				Name: fmt.Sprintf("release %s in %s %s", _r.Name(), _r.Destination().Type(), _r.Destination().Name()),
				Run: func(_ pool.StatusReporter) error {
					return r.configRepo.RenderForRelease(_r, helmfileArgs)
				},
			})
		}
	}

	if r.options.Scope != scope.Release && helmfileArgs.ArgocdMode {
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
			jobs = append(jobs, pool.Job{
				Name: fmt.Sprintf("global resources for %s %s", d.Type(), d.Name()),
				Run: func(_ pool.StatusReporter) error {
					return r.configRepo.RenderForDestination(d, helmfileArgs)
				},
			})
		}
	}

	return jobs, nil
}

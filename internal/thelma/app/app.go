// Package app contains logic for global/cross-cutting Thelma dependencies, such as configuration, logging support, and API client factories
package app

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/app/paths"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/app/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

// Options for a thelmaApp
type Options struct {
	Runner      shell.Runner
	StateLoader terra.StateLoader
}

func init() {
	logging.Bootstrap()
	seed.Rand()
}

// ThelmaApp holds references to global/cross-cutting dependencies for Thelma commands
type ThelmaApp interface {
	// Config returns configuration object for this ThelmaApp
	Config() config.Config
	// ShellRunner returns ShellRunner for this ThelmaApp
	ShellRunner() shell.Runner
	// Paths returns Paths for this ThelmaApp
	Paths() paths.Paths
	// Scratch returns the Scratch instance for this ThelmaApp
	Scratch() scratch.Scratch
	// State returns a new terra.State instance for this ThelmaApp
	State() (terra.State, error)
	// Close deletes local resources associated with this ThelmaApp, and should be called once before the program exits.
	Close() error
}

// New constructs a new ThelmaApp
func New(cfg config.Config, shellRunner shell.Runner, stateLoader terra.StateLoader) (ThelmaApp, error) {
	app := &thelmaApp{}
	app.config = cfg
	app.shellRunner = shellRunner
	app.stateLoader = stateLoader

	// Initialize paths
	_paths, err := paths.New(cfg)
	if err != nil {
		return nil, err
	}
	app.paths = _paths

	// Initialize scratch
	_scratch, err := scratch.NewScratch(cfg)
	if err != nil {
		return nil, err
	}
	app.scratch = _scratch

	return app, nil
}

type thelmaApp struct {
	config      config.Config
	shellRunner shell.Runner
	paths       paths.Paths
	scratch     scratch.Scratch
	stateLoader terra.StateLoader
}

func (t *thelmaApp) Config() config.Config {
	return t.config
}

func (t *thelmaApp) ShellRunner() shell.Runner {
	return t.shellRunner
}

func (t *thelmaApp) Paths() paths.Paths {
	return t.paths
}

func (t *thelmaApp) Scratch() scratch.Scratch {
	return t.scratch
}

func (t *thelmaApp) State() (terra.State, error) {
	return t.stateLoader.Load()
}

func (t *thelmaApp) Close() error {
	return t.scratch.Cleanup()
}

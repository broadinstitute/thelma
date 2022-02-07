package app

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/paths"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

// Options for a thelmaApp
type Options struct {
	Runner shell.Runner
}

// ThelmaApp holds references to global/cross-cutting dependencies for Thelma commands
type ThelmaApp interface {
	// Config returns configuration object for this ThelmaApp
	Config() config.Config
	// ShellRunner returns ShellRunner for this ThelmaApp
	ShellRunner() shell.Runner
	// Paths returns Paths for this ThelmaApp
	Paths() paths.Paths
	// Scratch returns a Scratch instance for this ThelmaApp
	Scratch() scratch.Scratch
	// Close deletes local resources associated with this ThelmaApp, and should be called once before the program exits.
	Close() error
}

// New construct a new App, given a Config
func New(cfg config.Config) (ThelmaApp, error) {
	return NewWithOptions(cfg, Options{})
}

// NewWithOptions construct a new App, given a Config & options
func NewWithOptions(cfg config.Config, options Options) (ThelmaApp, error) {
	app := &thelmaApp{}
	app.config = cfg

	// Initialize paths
	_paths, err := paths.New(cfg)
	if err != nil {
		return nil, err
	}
	app.paths = _paths

	// Initialize ShellRunner
	if options.Runner != nil {
		app.shellRunner = options.Runner
	} else {
		app.shellRunner = shell.NewRunner()
	}

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

func (t *thelmaApp) Close() error {
	return t.scratch.Cleanup()
}

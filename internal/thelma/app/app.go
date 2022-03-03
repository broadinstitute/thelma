package app

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/app/paths"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/app/seed"
	"github.com/broadinstitute/thelma/internal/thelma/gitops"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

// Options for a thelmaApp
type Options struct {
	Runner shell.Runner
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
	// TerraState returns a new terra.State instance for this ThelmaApp
	TerraState() (terra.State, error)
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

func (t *thelmaApp) TerraState() (terra.State, error) {
	return gitops.Load(t.config.Home(), t.shellRunner)
}

func (t *thelmaApp) Close() error {
	return t.scratch.Cleanup()
}

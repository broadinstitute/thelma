// Package app contains logic for global/cross-cutting Thelma dependencies, such as configuration, logging support, and API client factories
package app

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	_ "github.com/broadinstitute/thelma/internal/thelma/app/logging" // import logging for side effects (trigger bootstrapping)
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics"
	"github.com/broadinstitute/thelma/internal/thelma/app/paths"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/app/seed"
	"github.com/broadinstitute/thelma/internal/thelma/clients"
	"github.com/broadinstitute/thelma/internal/thelma/ops"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
)

// Options for a thelmaApp
type Options struct {
	Runner      shell.Runner
	StateLoader terra.StateLoader
}

func init() {
	seed.Rand()
}

// ThelmaApp holds references to global/cross-cutting dependencies for Thelma commands
type ThelmaApp interface {
	// Clients convenience constructors for clients used in Thelma commands
	Clients() clients.Clients
	// Config returns configuration object for this ThelmaApp
	Config() config.Config
	// Credentials returns credential manager object for this ThelmaApp
	Credentials() credentials.Credentials
	// AutoUpdate returns the installer for this ThelmaApp
	AutoUpdate() autoupdate.AutoUpdate
	// Ops returns the Ops interface for this ThelmaApp
	Ops() ops.Ops
	// Paths returns Paths for this ThelmaApp
	Paths() paths.Paths
	// Scratch returns the Scratch instance for this ThelmaApp
	Scratch() scratch.Scratch
	// ShellRunner returns ShellRunner for this ThelmaApp
	ShellRunner() shell.Runner
	// State returns a new terra.State instance for this ThelmaApp
	State() (terra.State, error)
	// StateLoader returns the terra.StateLoader instance for this ThelmaApp
	StateLoader() (terra.StateLoader, error)
	// Close deletes local resources associated with this ThelmaApp, and should be called once before the program exits.
	Close() error
}

// New constructs a new ThelmaApp
func New(cfg config.Config, creds credentials.Credentials, clients clients.Clients, installer autoupdate.AutoUpdate, scratch scratch.Scratch, shellRunner shell.Runner, stateLoader lazy.LazyE[terra.StateLoader], manageSingletons bool) (ThelmaApp, error) {
	app := &thelmaApp{}

	// Initialize paths
	_paths, err := paths.New(cfg)
	if err != nil {
		return nil, err
	}
	app.paths = _paths

	return &thelmaApp{
		clients:          clients,
		config:           cfg,
		credentials:      creds,
		installer:        installer,
		paths:            _paths,
		scratch:          scratch,
		shellRunner:      shellRunner,
		stateLoader:      stateLoader,
		manageSingletons: manageSingletons,
	}, nil
}

type thelmaApp struct {
	clients          clients.Clients
	config           config.Config
	credentials      credentials.Credentials
	installer        autoupdate.AutoUpdate
	paths            paths.Paths
	scratch          scratch.Scratch
	shellRunner      shell.Runner
	stateLoader      lazy.LazyE[terra.StateLoader]
	manageSingletons bool
}

func (t *thelmaApp) Clients() clients.Clients {
	return t.clients
}

func (t *thelmaApp) Config() config.Config {
	return t.config
}

func (t *thelmaApp) Credentials() credentials.Credentials {
	return t.credentials
}

func (t *thelmaApp) AutoUpdate() autoupdate.AutoUpdate {
	return t.installer
}

func (t *thelmaApp) Ops() ops.Ops {
	return ops.NewOps(t.clients)
}

func (t *thelmaApp) Paths() paths.Paths {
	return t.paths
}

func (t *thelmaApp) Scratch() scratch.Scratch {
	return t.scratch
}

func (t *thelmaApp) ShellRunner() shell.Runner {
	return t.shellRunner
}

func (t *thelmaApp) State() (terra.State, error) {
	s, err := t.StateLoader()
	if err != nil {
		return nil, err
	}
	return s.Load()
}

func (t *thelmaApp) StateLoader() (terra.StateLoader, error) {
	return t.stateLoader.Get()
}

func (t *thelmaApp) Close() error {
	if t.manageSingletons {
		if err := metrics.Push(); err != nil {
			log.Warn().Err(err).Msg("error pushing metrics")
		}
	}
	return t.scratch.Cleanup()
}

package builder

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate"
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics"
	"github.com/broadinstitute/thelma/internal/thelma/app/scratch"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"testing"

	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/clients"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	sherlockState "github.com/broadinstitute/thelma/internal/thelma/state/providers/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

// ThelmaBuilder is a utility for initializing new ThelmaApp instances
type ThelmaBuilder interface {
	// Build when first called, initializes a new ThelmaApp and saves it. Subsequent calls do nothing.
	Build() (app.ThelmaApp, error)
	// WithTestDefaults (FOR USE IN UNIT TESTS ONLY) causes app to be initialized with some settings that are useful
	// in testing (eg. ignore config file and environment variables when loading config).
	WithTestDefaults(t *testing.T) ThelmaBuilder
	// SetHome (FOR USE IN UNIT TESTS ONLY) sets the Thelma home directory to the given path.
	SetHome(string) ThelmaBuilder
	// SetConfigOverride (FOR USE IN UNIT TESTS ONLY) sets a configuration override for the Thelma app.
	SetConfigOverride(key string, value interface{}) ThelmaBuilder
	// SetConfigOption (FOR USE IN UNIT TESTS ONLY) customizes configuration behavior for the Thelma app. (see config.Load for more info)
	SetConfigOption(option config.Option) ThelmaBuilder
	// NoManageSingletons (FOR USE IN UNIT TESTS ONLY) prevent this builder from initializing singletons
	NoManageSingletons() ThelmaBuilder
	// SetShellRunner (FOR USE IN UNIT TESTS ONLY) sets the shell runner that the Thelma app should use.
	SetShellRunner(shell.Runner) ThelmaBuilder
	// UseCustomStateLoader (FOR USE IN UNIT TESTS ONLY) configures Thelma to use a custom state loader instead of a "real" terra.State
	UseCustomStateLoader(stateLoader terra.StateLoader) ThelmaBuilder
}

type thelmaBuilder struct {
	configOptions     []config.Option
	manageSingletons  bool
	shellRunner       shell.Runner
	customStateLoader terra.StateLoader
	rootDir           string
}

func NewBuilder() ThelmaBuilder {
	return &thelmaBuilder{
		configOptions:    []config.Option{},
		manageSingletons: true,
	}
}

func (b *thelmaBuilder) WithTestDefaults(t *testing.T) ThelmaBuilder {
	b.NoManageSingletons()

	// Set thelma root to empty temp directory
	b.SetRootDir(t.TempDir())

	// Set test-friendly configuration options
	b.SetConfigOption(config.WithTestDefaults(t))

	// Use mock shell runner
	b.SetShellRunner(shell.DefaultMockRunner())

	return b
}

func (b *thelmaBuilder) NoManageSingletons() ThelmaBuilder {
	b.manageSingletons = false
	return b
}

func (b *thelmaBuilder) SetHome(path string) ThelmaBuilder {
	b.SetConfigOverride(config.HomeKey, path)
	return b
}

func (b *thelmaBuilder) SetRootDir(dir string) ThelmaBuilder {
	b.rootDir = dir
	return b
}

func (b *thelmaBuilder) SetConfigOverride(key string, value interface{}) ThelmaBuilder {
	b.SetConfigOption(config.WithOverride(key, value))
	return b
}

func (b *thelmaBuilder) SetConfigOption(option config.Option) ThelmaBuilder {
	b.configOptions = append(b.configOptions, option)
	return b
}

func (b *thelmaBuilder) SetShellRunner(shellRunner shell.Runner) ThelmaBuilder {
	b.shellRunner = shellRunner
	return b
}

func (b *thelmaBuilder) UseCustomStateLoader(stateLoader terra.StateLoader) ThelmaBuilder {
	b.customStateLoader = stateLoader
	return b
}

func (b *thelmaBuilder) Build() (app.ThelmaApp, error) {
	rootDir := b.rootDir
	if rootDir == "" {
		rootDir = root.Lookup()
	}
	thelmaRoot := root.NewAt(rootDir)
	if err := thelmaRoot.CreateDirectories(); err != nil {
		return nil, err
	}

	// Initialize config
	var configOptions []config.Option
	configOptions = append(configOptions, config.WithThelmaRoot(thelmaRoot))
	configOptions = append(configOptions, b.configOptions...)

	cfg, err := config.Load(configOptions...)
	if err != nil {
		return nil, err
	}

	// Initialize logging
	if b.manageSingletons {
		if err := logging.Initialize(cfg, thelmaRoot); err != nil {
			return nil, err
		}
	}

	_credentials, err := credentials.New(cfg, thelmaRoot)
	if err != nil {
		return nil, err
	}

	shellRunner, err := b.buildShellRunner()
	if err != nil {
		return nil, err
	}

	// Initialize client factory
	_clients, err := clients.New(cfg, thelmaRoot, _credentials, shellRunner)
	if err != nil {
		return nil, err
	}

	// Initialize scratch
	_scratch, err := scratch.NewScratch(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize installer
	_installer, err := autoupdate.New(cfg, _clients.Google(), thelmaRoot, shellRunner, _scratch)
	if err != nil {
		return nil, err
	}

	// start backgrond update, if enabled
	if b.manageSingletons {
		if err = _installer.StartBackgroundUpdateIfEnabled(); err != nil {
			return nil, err
		}
	}

	// Initialize metrics
	if b.manageSingletons {
		iapToken, err := _clients.IAPToken()
		if err != nil {
			return nil, err
		}
		if err := metrics.Initialize(cfg, iapToken); err != nil {
			return nil, err
		}
	}

	stateLoader := b.buildStateLoader(cfg, _clients)

	// Initialize app
	return app.New(cfg, _credentials, _clients, _installer, _scratch, shellRunner, stateLoader, b.manageSingletons)
}

func (b *thelmaBuilder) buildShellRunner() (shell.Runner, error) {
	if b.shellRunner != nil {
		return b.shellRunner, nil
	}
	finder, err := toolbox.NewToolFinder()
	if err != nil {
		return nil, err
	}

	return shell.NewRunner(finder), nil
}

func (b *thelmaBuilder) buildStateLoader(cfg config.Config, clients clients.Clients) lazy.LazyE[terra.StateLoader] {
	if b.customStateLoader != nil {
		return lazy.NewLazyE(func() (terra.StateLoader, error) {
			return b.customStateLoader, nil
		})
	}

	return lazy.NewLazyE(func() (terra.StateLoader, error) {
		sherlock, err := clients.Sherlock()
		if err != nil {
			return nil, err
		}
		return sherlockState.NewStateLoader(cfg.Home(), sherlock), nil
	})
}

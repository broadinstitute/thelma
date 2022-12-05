package builder

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics"
	"testing"

	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/clients"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statebucket"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statefixtures"
	sherlockState "github.com/broadinstitute/thelma/internal/thelma/state/providers/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
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
	// UseStateFixture (FOR USE IN UNIT TESTS ONLY) configures Thelma to use a state fixture instead of a "real" terra.State
	UseStateFixture(name statefixtures.FixtureName, t *testing.T) ThelmaBuilder
}

type thelmaBuilder struct {
	configOptions    []config.Option
	manageSingletons bool
	shellRunner      shell.Runner
	stateFixture     struct {
		enabled bool
		name    statefixtures.FixtureName
		t       *testing.T
	}
	rootDir string
}

type stateLoaderType int

const (
	gitopsStateLoader stateLoaderType = iota + 1
	sherlockStateLoader
	undefinedStateLoader
)
const stateLoaderConfigPrefix string = "stateloader"

type StateLoaderConfig struct {
	Source string `default:"sherlock"`
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

	// Use state loader filled with fake/pre-populated data
	b.UseStateFixture(statefixtures.Default, t)

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

func (b *thelmaBuilder) UseStateFixture(name statefixtures.FixtureName, t *testing.T) ThelmaBuilder {
	b.stateFixture.enabled = true
	b.stateFixture.name = name
	b.stateFixture.t = t
	return b
}

func (b *thelmaBuilder) Build() (app.ThelmaApp, error) {
	rootDir := b.rootDir
	if rootDir == "" {
		rootDir = root.DefaultDir()
	}
	thelmaRoot := root.New(rootDir)
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

	// Initialize shell runner
	shellRunner := b.shellRunner
	if shellRunner == nil {
		shellRunner = shell.NewRunner()
	}

	// Initialize client factory
	_clients, err := clients.New(cfg, thelmaRoot, _credentials, shellRunner)
	if err != nil {
		return nil, err
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

	stateLoader, err := b.buildStateLoader(cfg, shellRunner, _clients)
	if err != nil {
		return nil, fmt.Errorf("error constructing state loader: %v", err)
	}

	// Initialize app
	return app.New(cfg, _credentials, _clients, shellRunner, stateLoader, b.manageSingletons)
}

func (b *thelmaBuilder) buildStateLoader(cfg config.Config, shellRunner shell.Runner, clients clients.Clients) (terra.StateLoader, error) {
	if b.stateFixture.enabled {
		return statefixtures.NewFakeStateLoader(b.stateFixture.name, b.stateFixture.t, cfg.Home(), shellRunner)
	}

	stateLoaderType, err := getStateLoaderType(cfg)
	if err != nil {
		return nil, err
	}

	if stateLoaderType == gitopsStateLoader {
		sb, err := statebucket.New(cfg, clients.Google())
		if err != nil {
			return nil, err
		}
		return gitops.NewStateLoader(cfg.Home(), shellRunner, sb), nil
	} else if stateLoaderType == sherlockStateLoader {
		sherlock, err := clients.Sherlock()
		if err != nil {
			return nil, err
		}
		return sherlockState.NewStateLoader(cfg.Home(), shellRunner, sherlock), nil
	}

	return nil, fmt.Errorf("received an undefined state loader source: valid options gitops or sherlock")
}

func getStateLoaderType(cfg config.Config) (stateLoaderType, error) {
	var stateLoaderConfig StateLoaderConfig
	if err := cfg.Unmarshal(stateLoaderConfigPrefix, &stateLoaderConfig); err != nil {
		return 0, err
	}

	log.Info().Msgf("State Source: %s", stateLoaderConfig.Source)

	switch stateLoaderConfig.Source {
	case "gitops":
		return gitopsStateLoader, nil
	case "sherlock":
		return sherlockStateLoader, nil
	default:
		return undefinedStateLoader, nil
	}
}

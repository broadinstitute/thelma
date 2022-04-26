package builder

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statefixtures"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"testing"
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
	// SetShellRunner (FOR USE IN UNIT TESTS ONLY) sets the shell runner that the Thelma app should use.
	SetShellRunner(shell.Runner) ThelmaBuilder
	// UseStateFixture (FOR USE IN UNIT TESTS ONLY) configures Thelma to use a state fixture instead of a "real" terra.State
	UseStateFixture(name statefixtures.FixtureName, t *testing.T) ThelmaBuilder
}

type thelmaBuilder struct {
	configOptions []config.Option
	shellRunner   shell.Runner
	stateFixture  struct {
		enabled bool
		name    statefixtures.FixtureName
		t       *testing.T
	}
	rootDir string
}

func NewBuilder() ThelmaBuilder {
	return &thelmaBuilder{
		configOptions: make([]config.Option, 0),
	}
}

func (b *thelmaBuilder) WithTestDefaults(t *testing.T) ThelmaBuilder {
	// Set thelma root to temp directory
	b.SetRootDir(t.TempDir())

	b.SetConfigOption(func(options config.Options) config.Options {
		// Ignore config file and environment when loading configuration
		options.ConfigFile = ""
		options.EnvPrefix = ""
		// Set THELMA_HOME to os tmp dir. Tests will sometimes override this setting with SetHome()
		options.Overrides[config.HomeKey] = t.TempDir()
		return options
	})

	// Use mock shell runner
	b.SetShellRunner(shell.DefaultMockRunner())

	// Use state loader filled with fake/pre-populated data
	b.UseStateFixture(statefixtures.Default, t)

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
	b.SetConfigOption(func(options config.Options) config.Options {
		options.Overrides[key] = value
		return options
	})
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

	// Initialize config
	var configOptions []config.Option
	configOptions = append(configOptions, config.WithThelmaRoot(thelmaRoot))
	configOptions = append(configOptions, b.configOptions...)

	cfg, err := config.Load(configOptions...)
	if err != nil {
		return nil, err
	}

	// Initialize logging
	if err := logging.InitializeLogging(cfg, thelmaRoot); err != nil {
		return nil, err
	}

	// Initialize shell runner
	shellRunner := b.shellRunner
	if shellRunner == nil {
		shellRunner = shell.NewRunner()
	}

	stateLoader, err := b.getStateLoader(cfg.Home(), shellRunner)
	if err != nil {
		return nil, fmt.Errorf("error constructing state loader: %v", err)
	}

	// Initialize app
	return app.New(cfg, shellRunner, stateLoader)
}

func (b *thelmaBuilder) getStateLoader(thelmaHome string, shellRunner shell.Runner) (terra.StateLoader, error) {
	if !b.stateFixture.enabled {
		return gitops.NewStateLoader(thelmaHome, shellRunner)
	}
	return statefixtures.NewFakeStateLoader(b.stateFixture.name, b.stateFixture.t, thelmaHome, shellRunner)
}

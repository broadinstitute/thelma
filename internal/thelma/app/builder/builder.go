package builder

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"os"
)

// ThelmaBuilder is a utility for initializing new ThelmaApp instances
type ThelmaBuilder interface {
	// Build when first called, initializes a new ThelmaApp and saves it. Subsequent calls do nothing.
	Build() (app.ThelmaApp, error)
	// WithTestDefaults (FOR USE IN UNIT TESTS ONLY) causes app to be initialized with some settings that are useful
	// in testing (eg. ignore config file and environment variables when loading config).
	// Panics if this app has already been initialized.
	WithTestDefaults() ThelmaBuilder
	// SetHome (FOR USE IN UNIT TESTS ONLY) sets the Thelma home directory to the given path.
	// Panics if this app has already been initialized.
	SetHome(string) ThelmaBuilder
	// SetConfigOverride (FOR USE IN UNIT TESTS ONLY) sets a configuration override for the Thelma app.
	// Panics if this app has already been initialized.
	SetConfigOverride(key string, value interface{}) ThelmaBuilder
	// SetConfigOption (FOR USE IN UNIT TESTS ONLY) customizes configuration behavior for the Thelma app. (see config.Load for more info)
	// Panics if this app has already been initialized.
	SetConfigOption(option config.Option) ThelmaBuilder
	// SetShellRunner (FOR USE IN UNIT TESTS ONLY) sets the shell runner that the Thelma app should use.
	// Panics if this app has already been initialized.
	SetShellRunner(shell.Runner) ThelmaBuilder
}

type thelmaBuilder struct {
	configOptions []config.Option
	shellRunner   shell.Runner
}

func NewBuilder() ThelmaBuilder {
	return &thelmaBuilder{
		configOptions: make([]config.Option, 0),
	}
}

func (t *thelmaBuilder) WithTestDefaults() ThelmaBuilder {
	t.SetConfigOption(func(options config.Options) config.Options {
		// Ignore config file and environment when loading configuration
		options.ConfigFile = ""
		options.EnvPrefix = ""
		// Set THELMA_HOME to os tmp dir. Tests will usually override this setting with SetHome()
		options.Overrides[config.HomeKey] = os.TempDir()
		return options
	})
	return t
}

func (t *thelmaBuilder) SetHome(path string) ThelmaBuilder {
	t.SetConfigOverride(config.HomeKey, path)
	return t
}

func (t *thelmaBuilder) SetConfigOverride(key string, value interface{}) ThelmaBuilder {
	t.SetConfigOption(func(options config.Options) config.Options {
		options.Overrides[key] = value
		return options
	})
	return t
}

func (t *thelmaBuilder) SetConfigOption(option config.Option) ThelmaBuilder {
	t.configOptions = append(t.configOptions, option)
	return t
}

func (t *thelmaBuilder) SetShellRunner(shellRunner shell.Runner) ThelmaBuilder {
	t.shellRunner = shellRunner
	return t
}

func (t *thelmaBuilder) Build() (app.ThelmaApp, error) {
	// Initialize config
	cfg, err := config.Load(t.configOptions...)
	if err != nil {
		return nil, err
	}

	// Initialize logging
	if err := logging.InitializeLogging(cfg); err != nil {
		return nil, err
	}

	// Initialize app
	return app.NewWithOptions(cfg, app.Options{Runner: t.shellRunner})
}

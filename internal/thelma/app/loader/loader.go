package loader

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog"
	"os"
)

// ThelmaLoader is a utility for initializing new ThelmaApp instances
type ThelmaLoader interface {
	// App Returns the initialized ThelmaApp.
	// Panics if app has not yet been initialized.
	App() app.ThelmaApp
	// Load when first called, initializes a new ThelmaApp and saves it. Subsequent calls do nothing.
	Load() error
	// Initialized returns true if Load() has been called successfully, false otherwise
	Initialized() bool
	// Close closes the App if one was initialized, otherwise does nothing
	Close() error
	// SetConfigOverride (FOR USE IN UNIT TESTS ONLY) sets a configuration override for the Thelma app.
	// Panics if this app has already been initialized.
	SetConfigOverride(string, interface{})
	// SetShellRunner (FOR USE IN UNIT TESTS ONLY) sets the shell runner that the Thelma app should use.
	// Panics if this app has already been initialized.
	SetShellRunner(shell.Runner)
}

type thelmaLoader struct {
	app             app.ThelmaApp
	configOverrides map[string]interface{}
	shellRunner     shell.Runner
}

func NewLoader() ThelmaLoader {
	return &thelmaLoader{
		app:             nil,
		configOverrides: make(map[string]interface{}),
	}
}

func (t *thelmaLoader) SetConfigOverride(key string, value interface{}) {
	if t.Initialized() {
		panic(fmt.Errorf("attempt to set config override after initialization: %s=%v", key, value))
	}

	t.configOverrides[key] = value
}

func (t *thelmaLoader) SetShellRunner(shellRunner shell.Runner) {
	if t.Initialized() {
		panic(fmt.Errorf("attempt to set shell runner after initialization"))
	}

	t.shellRunner = shellRunner
}

func (t *thelmaLoader) App() app.ThelmaApp {
	if !t.Initialized() {
		panic(fmt.Errorf("attempt to access App config before aclling Load()"))
	}

	return t.app
}

func (t *thelmaLoader) Load() error {
	if t.Initialized() {
		return nil
	}

	// Initialize config
	cfg, err := config.Load(t.configOverrides)
	if err != nil {
		return err
	}

	// Initialize app
	_app, err := app.NewWithOptions(cfg, app.Options{Runner: t.shellRunner})
	if err != nil {
		return err
	}

	// Set log level
	setLogLevel(_app.Config().LogLevel())

	t.app = _app

	return nil
}

func (t *thelmaLoader) Close() error {
	if !t.Initialized() {
		return nil
	}

	return t.app.Close()
}

func (t *thelmaLoader) Initialized() bool {
	return t.app != nil
}

// Adjust global logging verbosity based on Thelma config
func setLogLevel(levelStr string) {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to parse log level %q: %v", levelStr, err)
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	}
	zerolog.SetGlobalLevel(level)
}

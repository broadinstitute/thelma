package builder

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog"
	"os"
)

// ThelmaBuilder is a utility for initializing new ThelmaApp instances
type ThelmaBuilder interface {
	// App Returns the initialized ThelmaApp.
	// Panics if app has not yet been initialized.
	App() app.ThelmaApp
	// Build when first called, initializes a new ThelmaApp and saves it. Subsequent calls do nothing.
	Build() (app.ThelmaApp, error)
	// Close closes the App if one was initialized, otherwise does nothing
	Close() error
	// SetConfigOverride (FOR USE IN UNIT TESTS ONLY) sets a configuration override for the Thelma app.
	// Panics if this app has already been initialized.
	SetConfigOverride(string, interface{})
	// SetShellRunner (FOR USE IN UNIT TESTS ONLY) sets the shell runner that the Thelma app should use.
	// Panics if this app has already been initialized.
	SetShellRunner(shell.Runner)
}

type thelmaBuilder struct {
	app             app.ThelmaApp
	configOverrides map[string]interface{}
	shellRunner     shell.Runner
}

func NewBuilder() ThelmaBuilder {
	return &thelmaBuilder{
		app:             nil,
		configOverrides: make(map[string]interface{}),
	}
}

func (t *thelmaBuilder) SetConfigOverride(key string, value interface{}) {
	if t.initialized() {
		panic(fmt.Errorf("attempt to set config override after initialization: %s=%v", key, value))
	}

	t.configOverrides[key] = value
}

func (t *thelmaBuilder) SetShellRunner(shellRunner shell.Runner) {
	if t.initialized() {
		panic(fmt.Errorf("attempt to set shell runner after initialization"))
	}

	t.shellRunner = shellRunner
}

func (t *thelmaBuilder) App() app.ThelmaApp {
	if !t.initialized() {
		panic(fmt.Errorf("attempt to access App config before aclling Load()"))
	}

	return t.app
}

func (t *thelmaBuilder) Build() (app.ThelmaApp, error) {
	if t.initialized() {
		return nil, nil
	}

	// Initialize config
	cfg, err := config.Load(t.configOverrides)
	if err != nil {
		return nil, err
	}

	// Initialize app
	_app, err := app.NewWithOptions(cfg, app.Options{Runner: t.shellRunner})
	if err != nil {
		return nil, err
	}

	// Set log level
	setLogLevel(_app.Config().LogLevel())

	t.app = _app

	return nil, nil
}

func (t *thelmaBuilder) Close() error {
	if !t.initialized() {
		return nil
	}

	return t.app.Close()
}

// Returns true if app has been initialized
func (t *thelmaBuilder) initialized() bool {
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

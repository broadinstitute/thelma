package builder

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetConfigOverride(t *testing.T) {
	builder := NewBuilder().WithTestDefaults()
	fakeHome := t.TempDir()

	// add config overrides... later options should have higher precedence
	builder.SetConfigOption(func(options config.Options) config.Options {
		options.Overrides["foo.greeting"] = "hello"
		options.Overrides["foo.day"] = "Thursday"
		return options
	})
	builder.SetHome(fakeHome)
	builder.SetConfigOption(func(options config.Options) config.Options {
		options.Overrides["foo.greeting"] = "hi"
		return options
	})

	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}

	// make sure home is set to what we expect
	assert.Equal(t, fakeHome, builder.App().Config().Home())

	// make sure our overrides were set correctly as well
	type fooConfig struct {
		Greeting string
		Day      string
	}
	cfg := fooConfig{}
	err = builder.App().Config().Unmarshal("foo", &cfg)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "hi", cfg.Greeting)
	assert.Equal(t, "Thursday", cfg.Day)
}

func TestSetShellRunner(t *testing.T) {
	builder := NewBuilder().WithTestDefaults()

	runner := shell.DefaultMockRunner()
	builder.SetShellRunner(runner)

	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}

	assert.Same(t, runner, builder.App().ShellRunner())
}

func TestSettersPanicIfInitialized(t *testing.T) {
	builder := NewBuilder().WithTestDefaults()

	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}

	assert.Panics(t, func() {
		builder.SetShellRunner(shell.DefaultMockRunner())
	})
	assert.Panics(t, func() {
		builder.SetConfigOption(func(options config.Options) config.Options {
			// does nothing
			return options
		})
	})
}

func TestAppPanicsIfNotInitialized(t *testing.T) {
	assert.Panics(t, func() {
		NewBuilder().App()
	})
}

func TestCloseClosesApp(t *testing.T) {
	builder := NewBuilder().WithTestDefaults()
	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}
	dir, err := builder.App().Scratch().Mkdir("fake-scratch-dir")
	if !assert.NoError(t, err) {
		return
	}

	assert.DirExists(t, dir)
	err = builder.Close()
	if !assert.NoError(t, err) {
		return
	}
	assert.NoDirExists(t, dir)
}

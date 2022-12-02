package builder

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetConfigOverride(t *testing.T) {
	builder := NewBuilder().WithTestDefaults(t)
	fakeHome := t.TempDir()

	// add config overrides... later options should have higher precedence
	builder.SetConfigOption(func(options *config.Options) {
		options.Overrides["foo.greeting"] = "hello"
		options.Overrides["foo.day"] = "Thursday"
	})
	builder.SetHome(fakeHome)
	builder.SetConfigOption(func(options *config.Options) {
		options.Overrides["foo.greeting"] = "hi"
	})

	_app, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}
	t.Cleanup(func() {
		if err := _app.Close(); err != nil {
			t.Error(err)
		}
	})

	// make sure home is set to what we expect
	assert.Equal(t, fakeHome, _app.Config().Home())

	// make sure our overrides were set correctly as well
	type fooConfig struct {
		Greeting string
		Day      string
	}
	cfg := fooConfig{}
	err = _app.Config().Unmarshal("foo", &cfg)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "hi", cfg.Greeting)
	assert.Equal(t, "Thursday", cfg.Day)
}

func TestSetShellRunner(t *testing.T) {
	builder := NewBuilder().WithTestDefaults(t)

	runner := shell.DefaultMockRunner()
	builder.SetShellRunner(runner)

	_app, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}
	t.Cleanup(func() {
		if err := _app.Close(); err != nil {
			t.Error(err)
		}
	})

	assert.Same(t, runner, _app.ShellRunner())
}

package builder

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetConfigOverride(t *testing.T) {
	builder := NewBuilder()
	fakeHome := t.TempDir()
	builder.SetConfigOverride(config.Keys.Home, fakeHome)

	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, fakeHome, builder.App().Config().Home())
}

func TestSetShellRunner(t *testing.T) {
	runner := shell.DefaultMockRunner()
	builder := NewBuilder()
	builder.SetConfigOverride(config.Keys.Home, t.TempDir())
	builder.SetShellRunner(runner)

	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}

	assert.Same(t, runner, builder.App().ShellRunner())
}

func TestSettersPanicIfInitialized(t *testing.T) {
	builder := NewBuilder()
	builder.SetConfigOverride(config.Keys.Home, t.TempDir())
	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}

	assert.Panics(t, func() {
		builder.SetShellRunner(shell.DefaultMockRunner())
	})
	assert.Panics(t, func() {
		builder.SetConfigOverride("ignored", "does not matter")
	})
}

func TestAppPanicsIfNotInitialized(t *testing.T) {
	assert.Panics(t, func() {
		NewBuilder().App()
	})
}

func TestCloseClosesApp(t *testing.T) {
	builder := NewBuilder()
	builder.SetConfigOverride(config.Keys.Home, t.TempDir())
	_, err := builder.Build()
	if !assert.NoError(t, err) {
		return
	}
	dir, err := builder.App().Paths().CreateScratchDir("fake-scratch-dir")
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

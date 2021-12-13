package loader

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetConfigOverride(t *testing.T) {
	loader := NewLoader()
	fakeHome := t.TempDir()
	loader.SetConfigOverride(config.Keys.Home, fakeHome)

	if !assert.NoError(t, loader.Load()) {
		return
	}

	assert.Equal(t, fakeHome, loader.App().Config().Home())
}

func TestSetShellRunner(t *testing.T) {
	runner := shell.DefaultMockRunner()
	loader := NewLoader()
	loader.SetConfigOverride(config.Keys.Home, t.TempDir())
	loader.SetShellRunner(runner)

	if !assert.NoError(t, loader.Load()) {
		return
	}

	assert.Same(t, runner, loader.App().ShellRunner())
}

func TestSettersPanicIfInitialized(t *testing.T) {
	loader := NewLoader()
	loader.SetConfigOverride(config.Keys.Home, t.TempDir())
	if !assert.NoError(t, loader.Load()) {
		return
	}

	assert.Panics(t, func() {
		loader.SetShellRunner(shell.DefaultMockRunner())
	})
	assert.Panics(t, func() {
		loader.SetConfigOverride("ignored", "does not matter")
	})
}

func TestAppPanicsIfNotInitialized(t *testing.T) {
	assert.Panics(t, func() {
		NewLoader().App()
	})
}

func TestCloseClosesApp(t *testing.T) {
	loader := NewLoader()
	loader.SetConfigOverride(config.Keys.Home, t.TempDir())
	if !assert.NoError(t, loader.Load()) {
		return
	}
	dir, err := loader.App().Paths().CreateScratchDir("fake-scratch-dir")
	if !assert.NoError(t, err) {
		return
	}

	assert.DirExists(t, dir)
	err = loader.Close()
	if !assert.NoError(t, err) {
		return
	}
	assert.NoDirExists(t, dir)
}

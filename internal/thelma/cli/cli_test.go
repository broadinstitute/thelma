package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

// noop command for testing CLI building
func newFakeCommand(name string) ThelmaCommand {
	return &fakeCommand{name: name}
}

type fakeCommand struct {
	name         string
	preRunError  error
	preRunCount  int
	runError     error
	runCount     int
	postRunError error
	postRunCount int
}

func (n *fakeCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = n.name
	return
}

func (n *fakeCommand) PreRun(_ app.ThelmaApp, _ RunContext) error {
	n.preRunCount++
	return n.preRunError
}

func (n *fakeCommand) Run(_ app.ThelmaApp, _ RunContext) error {
	n.runCount++
	return n.runError
}

func (n *fakeCommand) PostRun(_ app.ThelmaApp, _ RunContext) error {
	n.postRunCount++
	return n.postRunError
}

func Test_NewPanicsMissingIntermediate(t *testing.T) {
	assert.Panics(t, func() {
		New(func(options *Options) {
			options.AddCommand("baz", newFakeCommand("baz"))
			options.AddCommand("foo bar", newFakeCommand("bar"))
		})
	}, "should panic if intermediate command is missing")
}

func Test_Execute_PreRunError(t *testing.T) {
	cmd := newFakeCommand("fake").(*fakeCommand)
	cmd.preRunError = errors.New("derp")
	_cli := New(func(options *Options) {
		options.AddCommand("fake", cmd)
		options.SetArgs([]string{"fake"})
	})
	err := _cli.Execute()
	assert.Error(t, err, "pre-run should return error")
	assert.Equal(t, "derp", err.Error(), "pre-run should return error")
	assert.Equal(t, 1, cmd.preRunCount)
	assert.Equal(t, 0, cmd.runCount, "run should not be called if pre run error")
	assert.Equal(t, 1, cmd.postRunCount, "post run should still be called if pre-run error")
}

func Test_Execute_PreRunErrorParent(t *testing.T) {
	parent := newFakeCommand("parent").(*fakeCommand)
	parent.preRunError = errors.New("derp")

	child := newFakeCommand("child").(*fakeCommand)

	_cli := New(func(options *Options) {
		options.AddCommand("parent", parent)
		options.AddCommand("parent child", child)
		options.SetArgs([]string{"parent", "child"})
	})

	err := _cli.Execute()
	assert.Error(t, err, "parent pre-run error should be returned")
	assert.Equal(t, "derp", err.Error(), "parent pre-run error should be returned")
	assert.Equal(t, 1, parent.preRunCount)
	assert.Equal(t, 0, child.preRunCount, "child pre-run should not be run")
	assert.Equal(t, 0, parent.runCount)
	assert.Equal(t, 1, parent.postRunCount, "parent post-run should still be run")
	assert.Equal(t, 1, child.postRunCount, "child post-run should still be run")
}

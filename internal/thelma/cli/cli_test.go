package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

// noop command for testing CLI building
func newNoopCommand() ThelmaCommand {
	return &noopCommand{}
}

type noopCommand struct{}

func (n noopCommand) ConfigureCobra(_ *cobra.Command) {
	return
}

func (n noopCommand) PreRun(_ app.ThelmaApp, _ RunContext) error {
	return nil
}

func (n noopCommand) Run(_ app.ThelmaApp, _ RunContext) error {
	return nil
}

func (n noopCommand) PostRun(_ app.ThelmaApp, _ RunContext) error {
	return nil
}

func Test_NewPanicsMissingIntermediate(t *testing.T) {
	assert.Panics(t, func() {
		New(func(options *Options) {
			options.AddCommand("baz", newNoopCommand())
			options.AddCommand("foo bar", newNoopCommand())
		})
	}, "should panic if intermediate command is missing")
}

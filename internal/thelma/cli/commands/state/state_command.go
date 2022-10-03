package states

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = "Tools for interacting with Thelma's internal state"

type stateCommand struct{}

func NewStateCommand() cli.ThelmaCommand {
	return &stateCommand{}
}

func (c stateCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "state [action]"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (c stateCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

func (c stateCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (c stateCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

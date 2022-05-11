package bee

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Tools for interacting with BEEs (Branch Engineering Environments)`

type beeCommand struct{}

func NewBeeCommand() cli.ThelmaCommand {
	return &beeCommand{}
}

func (cmd *beeCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "bee"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *beeCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *beeCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *beeCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

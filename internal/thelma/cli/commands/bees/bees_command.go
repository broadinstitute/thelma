package bees

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Bulk operations on BEEs (Branch Engineering Environments)`

type command struct{}

func NewBeesCommand() cli.ThelmaCommand {
	return &command{}
}

func (cmd *command) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "bees"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *command) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *command) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *command) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

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

func (v *beeCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "bee"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (v *beeCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (v *beeCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (v *beeCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

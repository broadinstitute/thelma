package auth

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Thelma authentication tools`

type authCommand struct{}

func NewAuthCommand() cli.ThelmaCommand {
	return &authCommand{}
}

func (cmd *authCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "auth"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *authCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *authCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *authCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

package auth

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Authentication tools`

type authCommand struct {
	echo bool
}

func NewAuthCommand() cli.ThelmaCommand {
	return &authCommand{}
}

func (cmd *authCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "auth"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
	cobraCommand.PersistentFlags().BoolVar(&cmd.echo, "echo", false, "Print credentials to STDOUT (be careful!)")
}

func (cmd *authCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *authCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *authCommand) PostRun(_ app.ThelmaApp, rc cli.RunContext) error {
	if !cmd.echo {
		rc.UnsetOutput()
	}
	return nil
}

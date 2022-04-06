package login

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/spf13/cobra"
)

const helpMessage = `Generate a new token for authenticating to ArgoCD`

type loginCommand struct{}

func NewArgoCDLoginCommand() cli.ThelmaCommand {
	return &loginCommand{}
}

func (cmd *loginCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "login"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *loginCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *loginCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	return argocd.Login(app.Config(), app.ShellRunner())
}

func (cmd *loginCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

package argocd

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Tools for interacting with ArgoCD`

type argocdCommand struct{}

func NewArgoCDCommand() cli.ThelmaCommand {
	return &argocdCommand{}
}

func (cmd *argocdCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "argocd"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *argocdCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *argocdCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *argocdCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

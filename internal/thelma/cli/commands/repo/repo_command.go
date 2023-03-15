package repo

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Tools for interacting with github repos`

type repoCommand struct{}

func NewRepoCommand() cli.ThelmaCommand {
	return &repoCommand{}
}

func (cmd *repoCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "repo"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *repoCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func (cmd *repoCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *repoCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

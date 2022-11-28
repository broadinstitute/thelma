package slack

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Tools for interacting with Broad's Slack instance'`

type slackCommand struct{}

func NewSlackCommand() cli.ThelmaCommand {
	return &slackCommand{}
}

func (cmd *slackCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "slack"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (cmd *slackCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *slackCommand) Run(_ app.ThelmaApp, _ cli.RunContext) error {
	panic("Run() is only executed for leaf commands")
}

func (cmd *slackCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

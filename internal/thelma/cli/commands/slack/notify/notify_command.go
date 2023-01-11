package notify

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `Send a message to a user on the Broad Institute Slack`

type options struct {
	userEmail string
	markdown  string
}

var flagNames = struct {
	userEmail string
	markdown  string
}{
	userEmail: "user-email",
	markdown:  "markdown",
}

type notifyCommand struct {
	options options
}

func NewSlackNotifyCommand() cli.ThelmaCommand {
	return &notifyCommand{}
}

func (cmd *notifyCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "notify [options]"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.userEmail, flagNames.userEmail, "u", "EMAIL", "Required. The Broad Institute email of the user to send the message to.")
	cobraCommand.Flags().StringVarP(&cmd.options.markdown, flagNames.markdown, "m", "TEXT", "Required. The markdown to send to the user.")
}

func (cmd *notifyCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	if !flags.Changed(flagNames.userEmail) {
		return fmt.Errorf("no user email specified; --%s is required", flagNames.userEmail)
	}
	if !flags.Changed(flagNames.markdown) {
		return fmt.Errorf("no markdown specified; --%s is required", flagNames.markdown)
	}

	return nil
}

func (cmd *notifyCommand) Run(app app.ThelmaApp, _ cli.RunContext) error {
	slack, err := app.Clients().Slack()
	if err != nil {
		return err
	}
	return slack.SendDirectMessage(cmd.options.userEmail, cmd.options.markdown)
}

func (cmd *notifyCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

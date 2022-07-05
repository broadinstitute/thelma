package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

type seedMeCommand struct {
	options options
}

func NewBeeSeedMeCommand() cli.ThelmaCommand {
	return &seedMeCommand{}
}

func (cmd *seedMeCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "me [options]"
	cobraCommand.Short = "Shortcut to register yourself with the BEE"
	cobraCommand.Long = `Register your current gcloud ADC with the BEE.

Equivalent to running "thelma bee seed --no-steps --step-6-extra-user use-adc"
See thelma bee seed's documentation for more information.

Examples (you'd need to set the --name of your environment):
  - To register your local ADC user as a Terra user in your BEE:
      $ thelma bee seed me
`

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "required; name of the BEE to seed")
	cobraCommand.Flags().BoolVarP(&cmd.options.force, flagNames.force, "f", false, "attempt to ignore errors during seeding")
}

func (cmd *seedMeCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	// validate --name
	if !flags.Changed(flagNames.name) {
		return fmt.Errorf("--%s is required", flagNames.name)
	}

	return nil
}

func (cmd *seedMeCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	cmd.options.noSteps = true
	cmd.options.step6ExtraUser = []string{"use-adc"}
	return (&seedCommand{options: cmd.options}).Run(app, ctx)
}

func (cmd *seedMeCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

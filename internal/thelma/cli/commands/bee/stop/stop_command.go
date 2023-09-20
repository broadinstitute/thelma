package stop

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const helpMessage = `Stop an existing BEE (in other words, make it offline and set replica counts to zero)

Examples:

# Stop an existing BEE
thelma bee stop --name=swat-grungy-puma
`

var flagNames = struct {
	name   string
	notify string
	sync   string
}{
	name:   "name",
	notify: "notify",
	sync:   "string",
}

type options struct {
	name string
	bee.StartStopOptions
}

type stopCommand struct {
	options options
}

func NewBeeStopCommand() cli.ThelmaCommand {
	return &stopCommand{}
}

func (cmd *stopCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "stop [options]"
	cobraCommand.Short = "Stop an existing BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to stop.")
	cobraCommand.Flags().BoolVar(&cmd.options.Notify, flagNames.notify, true, "If the BEE owner should be notified upon stop.")
	cobraCommand.Flags().BoolVar(&cmd.options.Sync, flagNames.sync, true, "If the BEE should be ArgoCD synced to immediately stop all chart instances.")
}

func (cmd *stopCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	if !flags.Changed(flagNames.name) {
		return errors.Errorf("no environment name specified; --%s is required", flagNames.name)
	}

	return nil
}

func (cmd *stopCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	_bee, err := bees.StartStopWith(cmd.options.name, true, cmd.options.StartStopOptions)
	if _bee != nil {
		ctx.SetOutput(views.DescribeBee(_bee))
	}
	return err
}

func (cmd *stopCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

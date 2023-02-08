package start

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/spf13/cobra"
)

const helpMessage = `Start an existing BEE (in other words, make it not offline and bring back the normal replica counts)

Examples:

# Start an existing BEE
thelma bee start --name=swat-grungy-puma
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

type startCommand struct {
	options options
}

func NewBeeStartCommand() cli.ThelmaCommand {
	return &startCommand{}
}

func (cmd *startCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "start [options]"
	cobraCommand.Short = "Start an existing BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to start.")
	cobraCommand.Flags().BoolVar(&cmd.options.Notify, flagNames.notify, true, "If the BEE owner should be notified upon start.")
	cobraCommand.Flags().BoolVar(&cmd.options.Sync, flagNames.sync, true, "If the BEE should be ArgoCD synced to immediately start all chart instances.")
}

func (cmd *startCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	if !flags.Changed(flagNames.name) {
		return fmt.Errorf("no environment name specified; --%s is required", flagNames.name)
	}

	return nil
}

func (cmd *startCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	_bee, err := bees.StartStopWith(cmd.options.name, false, cmd.options.StartStopOptions)
	if _bee != nil {
		ctx.SetOutput(views.DescribeBee(_bee))
	}
	return err
}

func (cmd *startCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

package list

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/views"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/terra/filter"
	"github.com/spf13/cobra"
)

const helpMessage = `List all BEE environments

Examples:

thelma bee list
`

type options struct {
	template string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	template string
}{
	template: "template",
}

type listCommand struct {
	options options
}

func NewBeeListCommand() cli.ThelmaCommand {
	return &listCommand{}
}

func (cmd *listCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "list"
	cobraCommand.Short = "List BEEs"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.template, flagNames.template, "t", "", "Only list BEEs created from the given template")
}

func (cmd *listCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	// nothing to do here
	return nil
}

func (cmd *listCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	state, err := app.TerraState()
	if err != nil {
		return err
	}

	// only show dynamic environments
	f := filter.Environments().HasLifecycle(terra.Dynamic)

	if cmd.options.template != "" {
		template, err := state.Environments().Get(cmd.options.template)
		if err != nil {
			return err
		}
		if template == nil {
			return fmt.Errorf("--%s: no template by the name %q exists", flagNames.template, cmd.options.template)
		}
		f = f.And(filter.Environments().HasTemplate(template))
	}

	envs, err := state.Environments().Filter(f)
	if err != nil {
		return err
	}

	rc.SetOutput(views.ForTerraEnvs(envs))

	return nil
}

func (cmd *listCommand) PostRun(app app.ThelmaApp, ctx cli.RunContext) error {
	// nothing to do here
	return nil
}

package list

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/spf13/cobra"
)

const helpMessage = `List all BEE environments

Examples:

thelma bee list
`

type options struct {
	template  string
	matchName string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	template  string
	matchName string
}{
	template:  "template",
	matchName: "match-name",
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
	cobraCommand.Flags().StringVarP(&cmd.options.matchName, flagNames.matchName, "m", "", "Only list BEEs with names that include the given substring")

}

func (cmd *listCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

func (cmd *listCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}

	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	var filters []terra.EnvironmentFilter

	if cmd.options.template != "" {
		template, err := state.Environments().Get(cmd.options.template)
		if err != nil {
			return err
		}
		if template == nil {
			return fmt.Errorf("--%s: no template by the name %q exists", flagNames.template, cmd.options.template)
		}
		filters = append(filters, filter.Environments().HasTemplate(template))
	}

	if cmd.options.matchName != "" {
		filters = append(filters, filter.Environments().NameIncludes(cmd.options.matchName))
	}

	matchingBees, err := bees.FilterBees(filter.Environments().And(filters...))
	if err != nil {
		return err
	}

	view := views.SummarizeBees(matchingBees)
	rc.SetOutput(view)

	return nil
}

func (cmd *listCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

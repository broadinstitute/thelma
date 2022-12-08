package list

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/filterflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/spf13/cobra"
)

const helpMessage = `List all BEE environments

Examples:

thelma bee list
`

type listCommand struct {
	fflags filterflags.FilterFlags
}

func NewBeeListCommand() cli.ThelmaCommand {
	return &listCommand{
		fflags: filterflags.NewFilterFlags(),
	}
}

func (cmd *listCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "list"
	cobraCommand.Short = "List BEEs"
	cobraCommand.Long = helpMessage

	cmd.fflags.AddFlags(cobraCommand)
}

func (cmd *listCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

func (cmd *listCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	beeFilter, err := cmd.fflags.GetFilter(app)
	if err != nil {
		return err
	}
	matchingBees, err := bees.FilterBees(beeFilter)
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

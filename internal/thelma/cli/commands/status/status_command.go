package status

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/common"
	"github.com/broadinstitute/thelma/internal/thelma/cli/selector"
	"github.com/spf13/cobra"
)

const helpMessage = `Report status information for a Terra service`

type statusCommand struct {
	selector *selector.Selector
}

func NewStatusCommand() cli.ThelmaCommand {
	return &statusCommand{
		selector: selector.NewSelector(),
	}
}

func (cmd *statusCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "status"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	// Release selector flags -- these flags determine which Argo apps will be synced
	cmd.selector.AddFlags(cobraCommand)
}

func (cmd *statusCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *statusCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	// compute selected releases
	state, err := app.State()
	if err != nil {
		return err
	}
	releases, err := cmd.selector.GetSelection(state, rc.CobraCommand().Flags(), rc.Args())
	if err != nil {
		return err
	}

	statusReader, err := app.Ops().Status()
	if err != nil {
		return err
	}
	statuses, err := statusReader.Statuses(releases)
	if err != nil {
		return err
	}
	rc.SetOutput(common.ReleaseMapToStructuredView(statuses))
	return nil
}

func (cmd *statusCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

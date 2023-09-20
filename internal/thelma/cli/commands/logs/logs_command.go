package logs

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/common"
	"github.com/broadinstitute/thelma/internal/thelma/cli/selector"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts/artifactsflags"
	"github.com/broadinstitute/thelma/internal/thelma/ops/logs"
	"github.com/spf13/cobra"
)

const helpMessage = `View container logs for Terra services`

type options struct {
	export bool
}

var flagNames = struct {
	export string
}{
	export: "export",
}

type logsCommand struct {
	artifactsFlags artifactsflags.ArtifactsFlags
	selector       *selector.Selector
	options        options
}

func NewLogsCommand() cli.ThelmaCommand {
	return &logsCommand{
		artifactsFlags: artifactsflags.NewArtifactsFlags(),
		selector:       selector.NewSelector(),
	}
}

func (cmd *logsCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "logs"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().BoolVarP(&cmd.options.export, flagNames.export, "x", false, "Export container logs to file")
	// Add artifacts flags
	cmd.artifactsFlags.AddFlags(cobraCommand)
	// Add release selector flags
	cmd.selector.AddFlags(cobraCommand)
}

func (cmd *logsCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *logsCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	// compute selected releases
	state, err := app.State()
	if err != nil {
		return err
	}

	selection, err := cmd.selector.GetSelection(state, rc.CobraCommand().Flags(), rc.Args())
	if err != nil {
		return err
	}

	_logs := app.Ops().Logs()

	if !cmd.options.export {
		if len(selection) != 1 {
			return fmt.Errorf("please specify exactly one chart release (matched %d)", len(selection))
		}
		return _logs.Logs(selection[0])
	}

	artifactsOptions, err := cmd.artifactsFlags.GetOptions()
	if err != nil {
		return err
	}

	locations, err := _logs.Export(selection, func(options *logs.ExportOptions) {
		options.Artifacts = artifactsOptions
	})
	if err != nil {
		return err
	}
	rc.SetOutput(common.ReleaseMapToStructuredView(locations))
	return nil
}

func (cmd *logsCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

package export

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/selector"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts/artifactsflags"
	"github.com/broadinstitute/thelma/internal/thelma/ops/logs"
	"github.com/spf13/cobra"
)

const helpMessage = `View container logs for Terra services`

type logsCommand struct {
	artifactsFlags artifactsflags.ArtifactsFlags
	selector       *selector.Selector
}

func NewLogsExportCommand() cli.ThelmaCommand {
	return &logsCommand{
		artifactsFlags: artifactsflags.NewArtifactsFlags(),
		selector: selector.NewSelector(func(options *selector.Options) {
			options.IncludeBulkFlags = false
			options.RequireDestination = true
		}),
	}
}

func (cmd *logsCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "logs"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	// Add artifacts flags
	cmd.artifactsFlags.AddFlags(cobraCommand)
	// Add release selctor flags
	cmd.selector.AddFlags(cobraCommand)
}

func (cmd *logsCommand) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *logsCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	artifactsOptions, err := cmd.artifactsFlags.GetOptions()
	if err != nil {
		return err
	}

	// compute selected releases
	state, err := app.State()
	if err != nil {
		return err
	}
	selection, err := cmd.selector.GetSelection(state, rc.CobraCommand().Flags(), rc.Args())
	if err != nil {
		return err
	}

	artifactsMgr := artifacts.NewManager(artifacts.ContainerLog, app.Clients().Google(), artifactsOptions)

	kubectl, err := app.Clients().Kubernetes().Kubectl()
	if err != nil {
		return err
	}

	exporter := logs.NewExporter(kubectl, artifactsMgr)

	locations, err := exporter.ExportLogs(selection.Releases)
	if err != nil {
		return err
	}
	rc.SetOutput(locations)
	return nil
}

func (cmd *logsCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

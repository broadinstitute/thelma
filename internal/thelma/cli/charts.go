package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/spf13/cobra"
)

const chartsHelpMessage = `Tools for interacting with Terra Helm charts`

type chartsCLI struct {
	cobraCommand *cobra.Command
	publishCLI   *chartsPublishCLI
	importCLI    *chartsImportCLI
}

func newChartsCLI(builder builder.ThelmaBuilder) *chartsCLI {
	publishCLI := newChartsPublishCLI(builder)
	importCLI := newChartsImportCLI(builder)

	cmd := &cobra.Command{
		Use:   "charts [action]",
		Short: chartsHelpMessage,
		Long:  chartsHelpMessage,
	}
	cmd.AddCommand(
		publishCLI.cobraCommand,
		importCLI.cobraCommand,
	)
	return &chartsCLI{
		cobraCommand: cmd,
		publishCLI:   publishCLI,
		importCLI:    importCLI,
	}
}

package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/loader"
	"github.com/spf13/cobra"
)

const chartsHelpMessage = `Tools for interacting with Terra Helm charts`

type chartsCLI struct {
	cobraCommand *cobra.Command
	publishCLI   *chartsPublishCLI
}

func newChartsCLI(loader loader.ThelmaLoader) *chartsCLI {
	publishCLI := newChartsPublishCLI(loader)
	importCLI := newChartsImportCLI(loader)

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
	}
}

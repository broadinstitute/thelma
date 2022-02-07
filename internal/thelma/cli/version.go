package cli

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"github.com/broadinstitute/thelma/internal/thelma/cli/printing"
	"github.com/spf13/cobra"
)

const versionHelpMessage = `Reports Thelma's version`

type versionCLI struct {
	cobraCommand *cobra.Command
}

func newVersionCLI(builder builder.ThelmaBuilder) *versionCLI {
	cmd := &cobra.Command{
		Use:   "version",
		Short: versionHelpMessage,
		Long:  versionHelpMessage,
	}

	printer := printing.NewPrinter()
	printer.AddFlags(cmd)

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("expected 0 arguments, got %d: %s", len(args), args)
		}

		if err := printer.VerifyFlags(); err != nil {
			return err
		}

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return printer.PrintOutput(version.GetManifest(), cmd.OutOrStdout())
	}

	return &versionCLI{
		cobraCommand: cmd,
	}
}

package update

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/spf13/cobra"
)

const helpMessage = `
Update Thelma

By default, updates to latest version of Thelma, but can
be run with --version to install a specific version if desired.
`

var flagNames = struct {
	version string
}{
	version: "version",
}

type updateCommand struct {
	version string
}

func NewUpdateCommand() cli.ThelmaCommand {
	return &updateCommand{}
}

func (cmd *updateCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "update"
	cobraCommand.Short = "Update Thelma"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.version, flagNames.version, "v", "", `Thelma semantic version or tag to update to (eg. "v1.2.3", "latest")`)
}

func (cmd *updateCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *updateCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	if rc.CobraCommand().Flags().Changed(flagNames.version) {
		return app.Installer().UpdateTo(cmd.version)
	}
	return app.Installer().Update()
}

func (cmd *updateCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

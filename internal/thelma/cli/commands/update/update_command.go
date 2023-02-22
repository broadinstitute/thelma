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
	version   string
	bootstrap string
}{
	version:   "version",
	bootstrap: "bootstrap",
}

type updateCommand struct {
	version   string
	bootstrap bool
}

func NewUpdateCommand() cli.ThelmaCommand {
	return &updateCommand{}
}

func (cmd *updateCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "update"
	cobraCommand.Short = "Update Thelma"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.version, flagNames.version, "v", "", `Thelma semantic version or tag to update to (eg. "v1.2.3", "latest")`)
	cobraCommand.Flags().BoolVar(&cmd.bootstrap, flagNames.bootstrap, false, `Configure a new installation of Thelma`)
}

func (cmd *updateCommand) PreRun(app app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *updateCommand) Run(app app.ThelmaApp, rc cli.RunContext) error {
	if err := cmd.update(app, rc); err != nil {
		return err
	}
	if cmd.bootstrap {
		return app.AutoUpdate().Bootstrap()
	}
	return nil
}

func (cmd *updateCommand) update(app app.ThelmaApp, rc cli.RunContext) error {
	if rc.CobraCommand().Flags().Changed(flagNames.version) {
		return app.AutoUpdate().UpdateTo(cmd.version)
	} else {
		// update to the latest version of the user's configured tag
		return app.AutoUpdate().Update()
	}
}

func (cmd *updateCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

package version

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const helpMessage = `Reports Thelma's version`

type versionCommand struct{}

func NewVersionCommand() cli.ThelmaCommand {
	return &versionCommand{}
}

func (v *versionCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "version"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage
}

func (v *versionCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	if len(ctx.Args()) != 0 {
		return errors.Errorf("expected 0 arguments, got: %v", ctx.Args())
	}
	return nil
}

func (v *versionCommand) Run(_ app.ThelmaApp, ctx cli.RunContext) error {
	ctx.SetOutput(version.GetManifest())
	return nil
}

func (v *versionCommand) PostRun(app app.ThelmaApp, ctx cli.RunContext) error {
	// nothing to do here
	return nil
}

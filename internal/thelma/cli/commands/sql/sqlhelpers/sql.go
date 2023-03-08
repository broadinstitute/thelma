package sqlhelpers

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/spf13/cobra"
)

// Command an interface for `thelma sql ...` commands.
type Command interface {
	ConfigureCobra(cobraCommand *cobra.Command)
	Run(conn api.Connection, app app.ThelmaApp, rc cli.RunContext) error
}

// AsThelmaCommand convert a sql.Command into a cli.ThelmaCommand
func AsThelmaCommand(c Command) cli.ThelmaCommand {
	return &command{child: c}
}

package init

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	sqlcli "github.com/broadinstitute/thelma/internal/thelma/cli/commands/sql/sqlhelpers"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/spf13/cobra"
)

var flagNames = struct {
	shell string
}{
	shell: "shell",
}

func NewSqlConnectCommand() cli.ThelmaCommand {
	return sqlcli.AsThelmaCommand(&connectCommand{})
}

type connectCommand struct {
	shell bool
}

func (c *connectCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Flags().BoolVar(&c.shell, flagNames.shell, false, "Run /bin/bash instead of psql/mysql")
}

func (c *connectCommand) Run(conn api.Connection, app app.ThelmaApp, rc cli.RunContext) error {
	conn.Options.Shell = c.shell
	return app.Ops().Sql().Connect(conn)
}

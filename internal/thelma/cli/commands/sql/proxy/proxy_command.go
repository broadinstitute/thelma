package proxy

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	sqlcli "github.com/broadinstitute/thelma/internal/thelma/cli/commands/sql/sqlhelpers"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/spf13/cobra"
)

func NewSqlProxyCommand() cli.ThelmaCommand {
	return sqlcli.AsThelmaCommand(&proxyCommand{})
}

type proxyCommand struct {
}

func (p *proxyCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	// no flags yet
}

func (p *proxyCommand) Run(instance api.Connection, app app.ThelmaApp, rc cli.RunContext) error {
	panic("TODO")
}

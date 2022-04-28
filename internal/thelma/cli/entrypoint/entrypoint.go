package entrypoint

import (
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth"
	auth_iap "github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth/iap"
	auth_vault "github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth/vault"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee"
	bee_create "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/create"
	bee_destroy "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/delete"
	bee_list "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/list"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts"
	charts_import "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/import"
	charts_publish "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/publish"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/render"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/version"
	"github.com/rs/zerolog/log"
	"os"
)

// Note: this code lives outside the `cli` package in order to avoid a dependency cycle (packages under `cli/commands` depend on the `cli` package)

// Execute main entrypoint for Thelma
func Execute() {
	_cli := cli.New(withCommands)

	if err := _cli.Execute(); err != nil {
		log.Error().Msgf("%v", err)
		os.Exit(1)
	}
}

func withCommands(opts *cli.Options) {
	opts.AddCommand("auth", auth.NewAuthCommand())
	opts.AddCommand("auth iap", auth_iap.NewAuthIAPCommand())
	opts.AddCommand("auth vault", auth_vault.NewAuthVaultCommand())

	opts.AddCommand("bee", bee.NewBeeCommand())
	opts.AddCommand("bee create", bee_create.NewBeeCreateCommand())
	opts.AddCommand("bee list", bee_list.NewBeeListCommand())
	opts.AddCommand("bee delete", bee_destroy.NewBeeDeleteCommand())

	opts.AddCommand("charts", charts.NewChartsCommand())
	opts.AddCommand("charts import", charts_import.NewChartsImportCommand())
	opts.AddCommand("charts publish", charts_publish.NewChartsPublishCommand())

	opts.AddCommand("render", render.NewRenderCommand())
	opts.AddCommand("version", version.NewVersionCommand())
}

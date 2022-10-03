package entrypoint

import (
	"os"

	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/argocd"
	argocd_sync "github.com/broadinstitute/thelma/internal/thelma/cli/commands/argocd/sync"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth"
	auth_argocd "github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth/argocd"
	auth_iap "github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth/iap"
	auth_vault "github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth/vault"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee"
	bee_create "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/create"
	bee_delete "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/delete"
	bee_describe "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/describe"
	bee_list "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/list"
	bee_pin "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/pin"
	bee_reset "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/reset"
	bee_seed "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed/seed"
	bee_unseed "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed/unseed"
	bee_unpin "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/unpin"
	bee_vars "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/vars"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts"
	charts_import "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/import"
	charts_publish "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/publish"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/render"
	states "github.com/broadinstitute/thelma/internal/thelma/cli/commands/state"
	state_export "github.com/broadinstitute/thelma/internal/thelma/cli/commands/state/export"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/version"
	"github.com/rs/zerolog/log"
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
	opts.AddCommand("argocd", argocd.NewArgoCDCommand())
	opts.AddCommand("argocd sync", argocd_sync.NewArgoCDSyncCommand())

	opts.AddCommand("auth", auth.NewAuthCommand())
	opts.AddCommand("auth argocd", auth_argocd.NewAuthArgoCDCommand())
	opts.AddCommand("auth iap", auth_iap.NewAuthIAPCommand())
	opts.AddCommand("auth vault", auth_vault.NewAuthVaultCommand())

	opts.AddCommand("bee", bee.NewBeeCommand())
	opts.AddCommand("bee create", bee_create.NewBeeCreateCommand())
	opts.AddCommand("bee delete", bee_delete.NewBeeDeleteCommand())
	opts.AddCommand("bee describe", bee_describe.NewBeeDescribeCommand())
	opts.AddCommand("bee list", bee_list.NewBeeListCommand())
	opts.AddCommand("bee pin", bee_pin.NewBeePinCommand())
	opts.AddCommand("bee reset", bee_reset.NewBeeResetCommand())
	opts.AddCommand("bee seed", bee_seed.NewBeeSeedCommand())
	opts.AddCommand("bee unseed", bee_unseed.NewBeeUnseedCommand())
	opts.AddCommand("bee unpin", bee_unpin.NewBeeUnpinCommand())
	opts.AddCommand("bee vars", bee_vars.NewBeeVarsCommand())

	opts.AddCommand("charts", charts.NewChartsCommand())
	opts.AddCommand("charts import", charts_import.NewChartsImportCommand())
	opts.AddCommand("charts publish", charts_publish.NewChartsPublishCommand())

	opts.AddCommand("render", render.NewRenderCommand())

	opts.AddCommand("state", states.NewStateCommand())
	opts.AddCommand("state export", state_export.NewStateExportCommand())

	opts.AddCommand("version", version.NewVersionCommand())
}

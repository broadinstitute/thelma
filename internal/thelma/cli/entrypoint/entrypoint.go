package entrypoint

import (
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/repo"
	repoCreate "github.com/broadinstitute/thelma/internal/thelma/cli/commands/repo/create"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/update"
	"github.com/rs/zerolog/log"
	"os"

	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/argocd"
	argocd_sync "github.com/broadinstitute/thelma/internal/thelma/cli/commands/argocd/sync"

	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth"
	auth_iap "github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth/iap"
	auth_vault "github.com/broadinstitute/thelma/internal/thelma/cli/commands/auth/vault"

	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee"
	bee_create "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/create"
	bee_delete "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/delete"
	bee_describe "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/describe"
	bee_list "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/list"
	bee_pin "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/pin"
	bee_provision "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/provision"
	bee_reset "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/reset"
	bee_seed "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed/seed"
	bee_unseed "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed/unseed"
	bee_start "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/start"
	bee_stop "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/stop"
	bee_sync "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/sync"
	bee_unpin "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/unpin"
	bee_vars "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/vars"

	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bees"
	bees_apply_schedule "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bees/apply_schedule"
	bees_delete "github.com/broadinstitute/thelma/internal/thelma/cli/commands/bees/delete"

	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts"
	charts_deploy "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/deploy"
	charts_import "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/import"
	charts_publish "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/publish"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/logs"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/render"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/slack"
	slack_notify "github.com/broadinstitute/thelma/internal/thelma/cli/commands/slack/notify"

	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/sql"
	sql_connect "github.com/broadinstitute/thelma/internal/thelma/cli/commands/sql/connect"
	sql_init "github.com/broadinstitute/thelma/internal/thelma/cli/commands/sql/init"
	sql_proxy "github.com/broadinstitute/thelma/internal/thelma/cli/commands/sql/proxy"

	states "github.com/broadinstitute/thelma/internal/thelma/cli/commands/state"
	state_export "github.com/broadinstitute/thelma/internal/thelma/cli/commands/state/export"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/status"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/version"
)

// Note: this code lives outside the `cli` package in order to avoid a dependency cycle (packages under `cli/commands` depend on the `cli` package)

// Execute main entrypoint for Thelma
func Execute() {
	_cli := cli.New(withCommands)

	if err := _cli.Execute(); err != nil {
		log.Error().Err(err).Send()
		os.Exit(1)
	}
}

func withCommands(opts *cli.Options) {
	opts.AddCommand("argocd", argocd.NewArgoCDCommand())
	opts.AddCommand("argocd sync", argocd_sync.NewArgoCDSyncCommand())

	opts.AddCommand("auth", auth.NewAuthCommand())
	opts.AddCommand("auth iap", auth_iap.NewAuthIAPCommand())
	opts.AddCommand("auth vault", auth_vault.NewAuthVaultCommand())

	opts.AddCommand("bee", bee.NewBeeCommand())
	opts.AddCommand("bee create", bee_create.NewBeeCreateCommand())
	opts.AddCommand("bee provision", bee_provision.NewBeeProvisionCommand())
	opts.AddCommand("bee delete", bee_delete.NewBeeDeleteCommand())
	opts.AddCommand("bee describe", bee_describe.NewBeeDescribeCommand())
	opts.AddCommand("bee list", bee_list.NewBeeListCommand())
	opts.AddCommand("bee pin", bee_pin.NewBeePinCommand())
	opts.AddCommand("bee reset", bee_reset.NewBeeResetCommand())
	opts.AddCommand("bee seed", bee_seed.NewBeeSeedCommand())
	opts.AddCommand("bee start", bee_start.NewBeeStartCommand())
	opts.AddCommand("bee stop", bee_stop.NewBeeStopCommand())
	opts.AddCommand("bee sync", bee_sync.NewBeeSyncCommand())
	opts.AddCommand("bee unseed", bee_unseed.NewBeeUnseedCommand())
	opts.AddCommand("bee unpin", bee_unpin.NewBeeUnpinCommand())
	opts.AddCommand("bee vars", bee_vars.NewBeeVarsCommand())

	opts.AddCommand("bees", bees.NewBeesCommand())
	opts.AddCommand("bees delete", bees_delete.NewBeesDeleteCommand())
	opts.AddCommand("bees apply-schedule", bees_apply_schedule.NewBeesApplyScheduleCommand())

	opts.AddCommand("charts", charts.NewChartsCommand())
	opts.AddCommand("charts import", charts_import.NewChartsImportCommand())
	opts.AddCommand("charts publish", charts_publish.NewChartsPublishCommand())
	opts.AddCommand("charts deploy", charts_deploy.NewChartsDeployCommand())

	opts.AddCommand("logs", logs.NewLogsCommand())

	opts.AddCommand("render", render.NewRenderCommand())
	opts.AddCommand("repo", repo.NewRepoCommand())
	opts.AddCommand("repo create", repoCreate.NewCreateCommand())

	opts.AddCommand("slack", slack.NewSlackCommand())
	opts.AddCommand("slack notify", slack_notify.NewSlackNotifyCommand())

	opts.AddCommand("sql", sql.NewSqlCommand())
	opts.AddCommand("sql connect", sql_connect.NewSqlConnectCommand())
	opts.AddCommand("sql init", sql_init.NewSqlInitCommand())
	opts.AddCommand("sql proxy", sql_proxy.NewSqlProxyCommand())

	opts.AddCommand("state", states.NewStateCommand())
	opts.AddCommand("state export", state_export.NewStateExportCommand())

	opts.AddCommand("status", status.NewStatusCommand())

	opts.AddCommand("update", update.NewUpdateCommand())

	opts.AddCommand("version", version.NewVersionCommand())
}

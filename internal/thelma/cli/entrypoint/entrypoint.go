package entrypoint

import (
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts"
	_import "github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/import"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/publish"
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
	opts.AddCommand("charts", charts.NewChartsCommand())
	opts.AddCommand("charts import", _import.NewChartsImportCommand())
	opts.AddCommand("charts publish", publish.NewChartsPublishCommand())
	opts.AddCommand("render", render.NewRenderCommand())
	opts.AddCommand("version", version.NewVersionCommand())
}

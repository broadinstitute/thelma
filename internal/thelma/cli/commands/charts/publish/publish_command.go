package publish

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Publishes Helm charts for Terra services`
const defaultBucketName = "terra-helm"

type options struct {
	chartDir   string
	bucketName string
	dryRun     bool
	charts     []string
}

var flagNames = struct {
	chartDir   string
	bucketName string
	dryRun     string
}{
	chartDir:   "chart-dir",
	bucketName: "bucket",
	dryRun:     "dry-run",
}

type publishCommand struct {
	options *options
}

func NewChartsPublishCommand() cli.ThelmaCommand {
	return &publishCommand{
		options: &options{},
	}
}

func (cmd *publishCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "publish [options] [CHART1] [CHART2] ..."
	cobraCommand.Short = "Publishes Helm charts"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVar(&cmd.options.chartDir, flagNames.chartDir, "path/to/charts", "Publish charts from custom directory")
	cobraCommand.Flags().StringVar(&cmd.options.bucketName, flagNames.bucketName, defaultBucketName, "Publish charts to custom GCS bucket")
	cobraCommand.Flags().BoolVarP(&cmd.options.dryRun, flagNames.dryRun, "n", false, "Dry run (don't actually update Helm repo)")
}

func (cmd *publishCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	cmd.options.charts = ctx.Args()

	if ctx.CobraCommand().Flags().Changed(flagNames.chartDir) {
		expanded, err := utils.ExpandAndVerifyExists(cmd.options.chartDir, "chart directory")
		if err != nil {
			return err
		}
		cmd.options.chartDir = expanded
	} else {
		cmd.options.chartDir = app.Paths().ChartsDir()
	}

	return nil
}

func (cmd *publishCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	published, err := publishCharts(cmd.options, app)
	if err != nil {
		return err
	}

	if cmd.options.dryRun {
		log.Info().Msgf("This is a dry run; would have released %d charts to %s", len(published), cmd.options.bucketName)
	} else {
		log.Info().Msgf("Released %d charts to %s", len(published), cmd.options.bucketName)
	}

	ctx.SetOutput(published)

	return nil
}

func (cmd *publishCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do
	return nil
}

func publishCharts(options *options, app app.ThelmaApp) ([]views.ChartRelease, error) {
	if len(options.charts) == 0 {
		log.Warn().Msgf("No charts specified; exiting")
		return []views.ChartRelease{}, nil
	}

	pb, err := builders.Publisher(app, options.bucketName, options.dryRun)
	if err != nil {
		return nil, err
	}
	defer pb.CloseWarn()
	publisher := pb.Publisher()

	_versions, err := gitops.NewVersions(app.Config().Home(), app.ShellRunner())
	if err != nil {
		return nil, err
	}

	chartsDir, err := source.NewChartsDir(options.chartDir, publisher, _versions, app.ShellRunner())
	if err != nil {
		return nil, err
	}

	chartVersions, err := chartsDir.Release(options.charts)
	if err != nil {
		return nil, err
	}

	// Collate version map into a slice of chart releases
	var view []views.ChartRelease
	for chartName, chartVersion := range chartVersions {
		view = append(view, views.ChartRelease{
			Name:    chartName,
			Version: chartVersion,
			Repo:    options.bucketName,
		})
	}
	views.SortChartReleases(view)
	return view, nil
}

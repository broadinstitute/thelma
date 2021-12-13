package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/loader"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/cli/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/printing"
	"github.com/broadinstitute/thelma/internal/thelma/cli/views"
	"github.com/broadinstitute/thelma/internal/thelma/gitops"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const chartsPublishHelpMessage = `Publishes Helm charts for Terra services`
const chartsPublishDefaultBucketName = "terra-helm"

type chartsPublishOptions struct {
	chartDir   string
	bucketName string
	dryRun     bool
	charts     []string
}

var chartsPublishFlagNames = struct {
	chartDir   string
	bucketName string
	dryRun     string
}{
	chartDir:   "chart-dir",
	bucketName: "bucket",
	dryRun:     "dry-run",
}

type chartsPublishCLI struct {
	loader       loader.ThelmaLoader
	cobraCommand *cobra.Command
	options      *chartsPublishOptions
}

func newChartsPublishCLI(loader loader.ThelmaLoader) *chartsPublishCLI {
	options := chartsPublishOptions{}

	cobraCommand := &cobra.Command{
		Use:   "publish [options] [CHART1] [CHART2] ...",
		Short: "Publishes Helm charts",
		Long:  chartsPublishHelpMessage,
	}

	cobraCommand.Flags().StringVar(&options.chartDir, chartsPublishFlagNames.chartDir, "path/to/charts", "Publish charts from custom directory")
	cobraCommand.Flags().StringVar(&options.bucketName, chartsPublishFlagNames.bucketName, chartsPublishDefaultBucketName, "Publish charts to custom GCS bucket")
	cobraCommand.Flags().BoolVarP(&options.dryRun, chartsPublishFlagNames.dryRun, "n", false, "Dry run (don't actually update Helm repo)")

	printer := printing.NewPrinter()
	printer.AddFlags(cobraCommand)

	cobraCommand.PreRunE = func(cmd *cobra.Command, args []string) error {
		options.charts = args

		if err := loader.Load(); err != nil {
			return err
		}

		if cmd.Flags().Changed(chartsPublishFlagNames.chartDir) {
			expanded, err := utils.ExpandAndVerifyExists(options.chartDir, "chart directory")
			if err != nil {
				return err
			}
			options.chartDir = expanded
		} else {
			options.chartDir = loader.App().Paths().DefaultChartSrcDir()
		}

		if err := printer.VerifyFlags(); err != nil {
			return err
		}

		return nil
	}

	cobraCommand.RunE = func(cmd *cobra.Command, args []string) error {
		published, err := publishCharts(&options, loader.App())
		if err != nil {
			return err
		}

		if options.dryRun {
			log.Info().Msgf("This is a dry run; would have released %d charts to %s", len(published), options.bucketName)
		} else {
			log.Info().Msgf("Released %d charts to %s", len(published), options.bucketName)
		}
		return printer.PrintOutput(published, cmd.OutOrStdout())
	}

	return &chartsPublishCLI{
		loader:       loader,
		cobraCommand: cobraCommand,
		options:      &options,
	}
}

func publishCharts(options *chartsPublishOptions, app app.ThelmaApp) ([]views.ChartRelease, error) {
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

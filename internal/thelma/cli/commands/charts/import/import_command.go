package _import

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/charts/mirror"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/views"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"path"
)

const helpMessage = `Imports charts from public Helm repositories into the terra-helm-third-party-repo`
const defaultBucketName = "terra-helm-thirdparty"
const defaultConfigFile = "third-party-charts.yaml"

type importCommand struct {
	options *options
}

type options struct {
	configFile string
	bucketName string
	dryRun     bool
}

var flagNames = struct {
	configFile string
	bucketName string
	dryRun     string
}{
	configFile: "config-from",
	bucketName: "bucket",
	dryRun:     "dry-run",
}

func NewChartsImportCommand() cli.ThelmaCommand {
	return &importCommand{
		options: &options{},
	}
}

func (cmd *importCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "import [options]"
	cobraCommand.Short = "Imports third-party Helm charts into the terra-helm-thirdparty repo"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVar(&cmd.options.configFile, flagNames.configFile, path.Join("$THELMA_HOME", "etc", defaultConfigFile), "Path to import config file")
	cobraCommand.Flags().StringVar(&cmd.options.bucketName, flagNames.bucketName, defaultBucketName, "Publish charts to custom GCS bucket")
	cobraCommand.Flags().BoolVarP(&cmd.options.dryRun, flagNames.dryRun, "n", false, "Dry run (don't actually update Helm repo)")
}

func (cmd *importCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	if len(ctx.Args()) != 0 {
		return errors.Errorf("expected no positional arguments, got %v", ctx.Args())
	}

	if ctx.CobraCommand().Flags().Changed(flagNames.configFile) {
		expanded, err := utils.ExpandAndVerifyExists(cmd.options.configFile, "configFile")
		if err != nil {
			return err
		}
		cmd.options.configFile = expanded
	} else {
		cmd.options.configFile = path.Join(app.Paths().EtcDir(), defaultConfigFile)
	}

	return nil
}

func (cmd *importCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	imported, err := importCharts(cmd.options, app)
	if err != nil {
		return err
	}

	if cmd.options.dryRun {
		log.Info().Msgf("This is a dry run; would have imported %d charts to %s", len(imported), cmd.options.bucketName)
	} else {
		log.Info().Msgf("Imported %d charts to %s", len(imported), cmd.options.bucketName)
	}

	ctx.SetOutput(imported)

	return nil
}

func (cmd *importCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do
	return nil
}

func importCharts(options *options, app app.ThelmaApp) ([]views.ChartRelease, error) {
	pb, err := builders.Publisher(app, options.bucketName, options.dryRun)
	if err != nil {
		return nil, err
	}
	defer pb.CloseWarn()

	_mirror, err := mirror.NewMirror(pb.Publisher(), app.ShellRunner(), options.configFile)
	if err != nil {
		return nil, err
	}

	chartDefns, err := _mirror.ImportToMirror()
	if err != nil {
		return nil, err
	}

	// convert result to view
	var result []views.ChartRelease
	for _, chartDefn := range chartDefns {
		result = append(result, views.ChartRelease{
			Name:    chartDefn.ChartName(),
			Version: chartDefn.Version,
			Repo:    chartDefn.RepoName(),
		})
	}
	views.SortChartReleases(result)
	return result, nil
}

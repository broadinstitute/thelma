package publish

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/charts/changedfiles"
	"github.com/broadinstitute/thelma/internal/thelma/charts/releaser"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/views"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Publishes Helm charts for Terra services

EXAMPLES

Publish in dry-run mode (this won't actually update any systems, and is safe to
run on your local machine):

  thelma charts publish --dry-run agora workspacemanager thurloe

Actually publish a list of charts to the terra-helm bucket and report new versions to Sherlock:

  thelma charts publish agora workspacemanager thurloe

Publish a list of charts from a file trigger:

  thelma charts publish --file-trigger ./list-of-updated-files.txt

  Note: A file trigger is text file containing a newline-separated list of files in the
  terra-helmfile repo that have changed. This is used to determine which charts need to
  be published, and is used in GitHub actions workflows to determine which charts were
  updated by a particular PR.

  All paths in the file trigger should be relative to the root of the terra-helmfile repo.
  Example:

    charts/agora/templates/deployment.yaml
    charts/thurloe/values.yaml
    helmfile.yaml
`
const defaultBucketName = "terra-helm"
const sherlockProdURL = "https://sherlock.dsp-devops.broadinstitute.org"
const sherlockDevURL = "https://sherlock-dev.dsp-devops.broadinstitute.org"

type options struct {
	chartDir         string
	bucketName       string
	dryRun           bool
	charts           []string
	sherlock         []string
	softFailSherlock []string
	description      string
	changedFilesList string
}

var flagNames = struct {
	chartDir         string
	bucketName       string
	dryRun           string
	sherlock         string
	softFailSherlock string
	description      string
	fileTrigger      string
}{
	chartDir:         "chart-dir",
	bucketName:       "bucket",
	dryRun:           "dry-run",
	sherlock:         "sherlock",
	softFailSherlock: "soft-fail-sherlock",
	description:      "description",
	fileTrigger:      changedfiles.FlagName,
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
	cobraCommand.Flags().BoolVarP(&cmd.options.dryRun, flagNames.dryRun, "n", false, "Dry run (don't actually update Helm repo or release to any versioning systems)")
	cobraCommand.Flags().StringSliceVar(&cmd.options.sherlock, flagNames.sherlock, []string{sherlockProdURL}, "Sherlock servers to use as versioning systems to release to")
	cobraCommand.Flags().StringSliceVar(&cmd.options.softFailSherlock, flagNames.softFailSherlock, []string{sherlockDevURL}, "Sherlock server to use as versioning systems to release to, always using soft-fail behavior")
	cobraCommand.Flags().StringVarP(&cmd.options.description, flagNames.description, "d", "", "The description to use for these version bumps on any Sherlock versioning systems")
	cobraCommand.Flags().StringVarP(&cmd.options.changedFilesList, flagNames.fileTrigger, "f", "", "Path to a file trigger (see --help for more info)")
}

func (cmd *publishCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	cmd.options.charts = ctx.Args()
	if cmd.options.changedFilesList != "" {
		state, err := app.State()
		if err != nil {
			return err
		}
		chartsDir, err := source.NewChartsDir(app.Paths().ChartsDir(), app.ShellRunner())
		if err != nil {
			return err
		}
		changedFiles := changedfiles.New(chartsDir, state)
		chartsToPublish, err := changedFiles.ChartList(cmd.options.changedFilesList)
		if err != nil {
			return errors.Errorf("error building chart list from file trigger: %v", err)
		}
		cmd.options.charts = append(cmd.options.charts, chartsToPublish...)
	}

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

// publishCharts publishes the given charts and any transitive dependencies they have.
//
// During chart publishing, we use `helm dependency update` with `--skip-refresh` to save time.
// This requires the Helm repositories and their indexes to already exist, so we borrow
// Helmfile's `helmfile repos` capability to do this based on the same Helm repository
// configuration used for rendering manifests.
func publishCharts(options *options, app app.ThelmaApp) ([]views.ChartRelease, error) {
	if len(options.charts) == 0 {
		log.Warn().Msgf("No charts specified; exiting")
		return []views.ChartRelease{}, nil
	}

	stubHelmfileOptions := helmfile.Options{
		ThelmaHome:  app.Config().Home(),
		ShellRunner: app.ShellRunner(),
	}
	if err := helmfile.NewConfigRepo(stubHelmfileOptions).HelmUpdate(); err != nil {
		return nil, errors.Errorf("error using helmfile for `helmfile repos`: %v", err)
	}

	pb, err := builders.Publisher(app, options.bucketName, options.dryRun)
	if err != nil {
		return nil, err
	}
	defer pb.CloseWarn()
	publisher := pb.Publisher()

	updater := &releaser.DeployedVersionUpdater{}
	// If we're dry-running, the updater will be empty so we don't mutate anything.
	if !options.dryRun {
		if len(options.sherlock) > 0 || len(options.softFailSherlock) > 0 {
			for _, sherlockURL := range options.sherlock {
				if sherlockURL != "" {
					client, err := app.Clients().Sherlock(func(options *sherlock.Options) {
						options.Addr = sherlockURL
					})
					if err != nil {
						return nil, err
					}
					updater.SherlockUpdaters = append(updater.SherlockUpdaters, client)
				}
			}
			for _, sherlockURL := range options.softFailSherlock {
				if sherlockURL != "" {
					client, err := app.Clients().Sherlock(func(options *sherlock.Options) {
						options.Addr = sherlockURL
					})
					if err != nil {
						return nil, err
					}
					updater.SoftFailSherlockUpdaters = append(updater.SoftFailSherlockUpdaters, client)
				}
			}
		}
	}

	chartsDir, err := source.NewChartsDir(options.chartDir, app.ShellRunner())
	if err != nil {
		return nil, err
	}

	state, err := app.State()
	if err != nil {
		return nil, err
	}
	syncer := releaser.NewPostUpdateSyncer(app.Ops().Sync, state, options.dryRun)

	chartReleaser := releaser.NewChartReleaser(chartsDir, publisher, updater, syncer)

	chartVersions, err := chartReleaser.Release(options.charts, options.description)
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

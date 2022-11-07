package publish

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/views"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

const helpMessage = `Publishes Helm charts for Terra services`
const defaultBucketName = "terra-helm"
const iapIdTokenEnvironmentVariable = "THELMA_IAP_ID_TOKEN"
const sherlockProdURL = "https://sherlock.dsp-devops.broadinstitute.org"
const sherlockDevURL = "https://sherlock-dev.dsp-devops.broadinstitute.org"

type options struct {
	chartDir         string
	bucketName       string
	dryRun           bool
	charts           []string
	gitops           bool
	sherlock         []string
	softFailSherlock []string
	description      string
}

var flagNames = struct {
	chartDir         string
	bucketName       string
	dryRun           string
	gitops           string
	sherlock         string
	softFailSherlock string
	description      string
}{
	chartDir:         "chart-dir",
	bucketName:       "bucket",
	dryRun:           "dry-run",
	gitops:           "gitops",
	sherlock:         "sherlock",
	softFailSherlock: "soft-fail-sherlock",
	description:      "description",
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
	cobraCommand.Flags().BoolVar(&cmd.options.gitops, flagNames.gitops, true, "Use terra-helmfile gitops as one of the versioning systems to release to")
	cobraCommand.Flags().StringSliceVar(&cmd.options.sherlock, flagNames.sherlock, []string{sherlockProdURL}, "Sherlock servers to use as versioning systems to release to")
	cobraCommand.Flags().StringSliceVar(&cmd.options.softFailSherlock, flagNames.softFailSherlock, []string{sherlockDevURL}, "Sherlock server to use as versioning systems to release to, always using soft-fail behavior")
	cobraCommand.Flags().StringVarP(&cmd.options.description, flagNames.description, "d", "", "The description to use for these version bumps on any Sherlock versioning systems")
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
		return nil, fmt.Errorf("error using helmfile for `helmfile repos`: %v", err)
	}

	pb, err := builders.Publisher(app, options.bucketName, options.dryRun)
	if err != nil {
		return nil, err
	}
	defer pb.CloseWarn()
	publisher := pb.Publisher()

	autoreleaser := &source.AutoReleaser{}
	if options.gitops {
		gitopsVersions, err := gitops.NewVersions(app.Config().Home(), app.ShellRunner())
		if err != nil {
			return nil, err
		}
		autoreleaser.GitopsUpdaters = []gitops.Versions{gitopsVersions}
	}
	if len(options.sherlock) > 0 || len(options.softFailSherlock) > 0 {
		iapIdToken, err := getIapToken(app)
		if err != nil {
			return nil, err
		}
		for _, sherlockURL := range options.sherlock {
			client, err := sherlock.NewWithHostnameOverride(sherlockURL, iapIdToken)
			if err != nil {
				return nil, err
			}
			autoreleaser.SherlockUpdaters = append(autoreleaser.SherlockUpdaters, client)
		}
		for _, sherlockURL := range options.softFailSherlock {
			client, err := sherlock.NewWithHostnameOverride(sherlockURL, iapIdToken)
			if err != nil {
				return nil, err
			}
			autoreleaser.SoftFailSherlockUpdaters = append(autoreleaser.SoftFailSherlockUpdaters, client)
		}
	}

	chartsDir, err := source.NewChartsDir(options.chartDir, publisher, app.ShellRunner(), autoreleaser)
	if err != nil {
		return nil, err
	}

	chartVersions, err := chartsDir.PublishAndRelease(options.charts, options.description)
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

// getIapToken tries to read the IAP ID token itself from THELMA_IAP_ID_TOKEN, and if that fails, it tries to call
// Thelma's overall TokenProvider mechanism, which will load Vault auth and work with either Workload Identity or
// browser-based authentication.
// Note that this environment variable is only respected for this CLI command--other commands don't have the same
// shortcut behavior.
func getIapToken(app app.ThelmaApp) (string, error) {
	if token, found := os.LookupEnv(iapIdTokenEnvironmentVariable); found {
		return token, nil
	} else {
		return app.Clients().IAPToken()
	}
}

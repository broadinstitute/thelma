package deploy

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/charts/deploy"
	"github.com/broadinstitute/thelma/internal/thelma/charts/releaser"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/charts/views"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
)

const helpMessage = `Deploys Helm charts for Terra services

This command is intended to be run in GitHub Actions after:

1. "thelma charts publish" is used to package and upload 
   new versions of the charts to the terra-helm bucket

2. "git tag" and "git push" are run to create tags for the newly published
   chart versions

For example:

  # Publish new versions of charts that were updated by the merged PR
  thelma charts publish \
    --changed-files-list=changed-files.txt \
    --output-file=versions.yaml

  # Github Actions - create and push git tags for the newly-published chart
  # versions (omitted)
  ...

  # Deploy the updated charts to the dev environment
  thelma charts deploy --chart-versions=versions.yaml

IDENTIFYING TARGET CHART RELEASES

For each chart it is deploying, this command behaves as follows:

1. If there is a .autorelease.yaml file in the chart's source directory, it
   will deploy and sync the chart to all target releases specified in the
   file.
2. Else it will check to see if the chart has a chart release in Terra's dev
   environment; if so, it will deploy and sync the chart to the dev environment
   chart release.

EXAMPLES

Deploy the versions of the charts specified in a versions.yaml file produced by
the "thelma charts publish" command": 

  thelma charts deploy --versions-file=versions.yaml

Deploy in dry-run mode (this won't actually update any systems, and is safe
to run on your local machine).

  thelma charts deploy --versions-file=versions.yaml --dry-run

Deploy the leonardo and sam charts, pulling the desired version from Chart.yaml
files in the THELMA_HOME / --chart-dir directory:

  thelma charts deploy --dry-run leonardo sam

`

const sherlockProdURL = "https://sherlock.dsp-devops.broadinstitute.org"
const sherlockDevURL = "https://sherlock-dev.dsp-devops.broadinstitute.org"

type options struct {
	versionsFile      string
	chartDir          string
	dryRun            bool
	charts            []string
	sherlock          []string
	softFailSherlock  []string
	description       string
	ignoreSyncFailure bool
}

var flagNames = struct {
	versionsFile      string
	chartDir          string
	dryRun            string
	sherlock          string
	softFailSherlock  string
	description       string
	ignoreSyncFailure string
}{
	versionsFile:      "versions-file",
	chartDir:          "chart-dir",
	dryRun:            "dry-run",
	sherlock:          "sherlock",
	softFailSherlock:  "soft-fail-sherlock",
	description:       "description",
	ignoreSyncFailure: "ignore-sync-failure",
}

type deployCommand struct {
	options *options
}

func NewChartsDeployCommand() cli.ThelmaCommand {
	return &deployCommand{
		options: &options{},
	}
}

func (cmd *deployCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "deploy [options] [CHART1] [CHART2] ..."
	cobraCommand.Short = "Deploys Helm charts to dev"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.versionsFile, flagNames.versionsFile, "f", "", "Path to YAML-formated versions file produced by `thelma charts publish`")
	cobraCommand.Flags().StringVar(&cmd.options.chartDir, flagNames.chartDir, "path/to/charts", "Publish charts from custom directory")
	cobraCommand.Flags().BoolVarP(&cmd.options.dryRun, flagNames.dryRun, "n", false, "Dry run (don't actually update Helm repo or release to any versioning systems)")
	cobraCommand.Flags().StringSliceVar(&cmd.options.sherlock, flagNames.sherlock, []string{sherlockProdURL}, "Sherlock servers to use as versioning systems to release to")
	cobraCommand.Flags().StringSliceVar(&cmd.options.softFailSherlock, flagNames.softFailSherlock, []string{sherlockDevURL}, "Sherlock server to use as versioning systems to release to, always using soft-fail behavior")
	cobraCommand.Flags().StringVarP(&cmd.options.description, flagNames.description, "d", "", "The description to use for these version bumps on any Sherlock versioning systems")
	cobraCommand.Flags().BoolVar(&cmd.options.ignoreSyncFailure, flagNames.ignoreSyncFailure, true, "Ignore ArgoCD sync failures")
}

func (cmd *deployCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	// only one of --versions-file or pos args can be specified
	hasArgs := len(ctx.Args()) > 0
	hasVersionsFile := ctx.CobraCommand().Flags().Changed(flagNames.versionsFile)

	if hasArgs && hasVersionsFile {
		return errors.New("either chart names or --versions-file can be specified, but not both")
	}
	if !hasArgs && !hasVersionsFile {
		return errors.New(" chart names or --versions-file must be specified")
	}

	// verify / set default for chart dir
	if ctx.CobraCommand().Flags().Changed(flagNames.chartDir) {
		expanded, err := utils.ExpandAndVerifyExists(cmd.options.chartDir, "chart directory")
		if err != nil {
			return err
		}
		cmd.options.chartDir = expanded
	} else {
		cmd.options.chartDir = app.Paths().ChartsDir()
	}

	// verify versions file exists
	if ctx.CobraCommand().Flags().Changed(flagNames.versionsFile) {
		expanded, err := utils.ExpandAndVerifyExists(cmd.options.versionsFile, "versions file")
		if err != nil {
			return err
		}
		cmd.options.versionsFile = expanded
	}

	return nil
}

func (cmd *deployCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	chartsDir, err := source.NewChartsDir(cmd.options.chartDir, app.ShellRunner())
	if err != nil {
		return err
	}

	chartVersions, err := cmd.parseChartVersions(ctx, chartsDir)
	if err != nil {
		return err
	}

	if len(chartVersions) == 0 {
		log.Warn().Msgf("No charts to deploy (is the versions file empty?)")
		return nil
	}

	state, err := app.State()
	if err != nil {
		return err
	}

	updater, err := cmd.buildUpdater(app)
	if err != nil {
		return err
	}

	deployer, err := deploy.New(chartsDir, updater, app.Ops().Sync, state, deploy.Options{
		DryRun:            cmd.options.dryRun,
		IgnoreSyncFailure: cmd.options.ignoreSyncFailure,
	})
	if err != nil {
		return err
	}

	return deployer.Deploy(chartVersions, cmd.options.description)
}

func (cmd *deployCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do
	return nil
}

// parseChartVersions parse chart versions based on CLI args
func (cmd *deployCommand) parseChartVersions(ctx cli.RunContext, chartsDir source.ChartsDir) (map[string]releaser.VersionPair, error) {
	if ctx.CobraCommand().Flags().Changed(flagNames.versionsFile) {
		return parseChartVersions(cmd.options.versionsFile)
	} else {
		return loadChartVersionsFromSourceDir(ctx.Args(), chartsDir)
	}
}

// loadChartVersionsFromSourceDir load chart versions form charts directory
func loadChartVersionsFromSourceDir(chartNames []string, sourceDir source.ChartsDir) (map[string]releaser.VersionPair, error) {
	charts, err := sourceDir.GetCharts(chartNames...)
	if err != nil {
		return nil, errors.Errorf("failed to load charts from %s: %v", sourceDir.Path(), err)
	}

	versions := make(map[string]releaser.VersionPair)
	for _, chart := range charts {
		versions[chart.Name()] = releaser.VersionPair{
			NewVersion:   chart.ManifestVersion(),
			PriorVersion: "", // no way to know the prior version without a chart versions file
		}
	}

	return versions, nil
}

// parseChartVersions parse chart versions from a versions file
func parseChartVersions(versionsFile string) (map[string]releaser.VersionPair, error) {
	var view []views.ChartRelease

	content, err := os.ReadFile(versionsFile)
	if err != nil {
		return nil, errors.Errorf("error reading file %s: %v", versionsFile, err)
	}
	if err = yaml.Unmarshal(content, view); err != nil {
		return nil, errors.Errorf("error parsing versions file %s: %v", versionsFile, err)
	}

	versions := make(map[string]releaser.VersionPair)
	for _, chart := range view {
		versions[chart.Name] = releaser.VersionPair{
			NewVersion:   chart.Version,
			PriorVersion: chart.PriorVersion,
		}
	}

	return versions, nil
}

func (cmd *deployCommand) buildUpdater(app app.ThelmaApp) (deploy.DeployedVersionUpdater, error) {
	opts := cmd.options

	var updater deploy.DeployedVersionUpdater

	// If we're dry-running, the updater will be empty so we don't mutate anything.
	if opts.dryRun {
		return updater, nil
	}

	if len(opts.sherlock) > 0 || len(opts.softFailSherlock) > 0 {
		for _, sherlockURL := range opts.sherlock {
			if sherlockURL != "" {
				client, err := app.Clients().Sherlock(func(options *sherlock.Options) {
					options.Addr = sherlockURL
				})
				if err != nil {
					return updater, err
				}
				updater.SherlockUpdaters = append(updater.SherlockUpdaters, client)
			}
		}
		for _, sherlockURL := range opts.softFailSherlock {
			if sherlockURL != "" {
				client, err := app.Clients().Sherlock(func(options *sherlock.Options) {
					options.Addr = sherlockURL
				})
				if err != nil {
					return updater, err
				}
				updater.SoftFailSherlockUpdaters = append(updater.SoftFailSherlockUpdaters, client)
			}
		}
	}

	return updater, nil
}

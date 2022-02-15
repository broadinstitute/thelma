package render

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/render"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
	"github.com/broadinstitute/thelma/internal/thelma/render/resolver"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"path"
	"path/filepath"
)

const helpMessage = `Renders Terra Kubernetes manifests

Examples:

# Render all manifests for all Terra services in all environments
thelma render

# Render manifests for all Terra services in the dev environment
thelma render -e dev

# Render manifests for the cromwell service in the alpha environment
thelma render -e alpha -a cromwell

# Render manifests for the cromwell service in the alpha environment,
# overriding app and chart version
thelma render -e alpha -a cromwell --chart-version="~> 0.8" --app-version="53-9b11416"

# Render manifests from a local copy of a chart
thelma render -e alpha -a cromwell --chart-dir=../terra-helm/charts

# Render manifests, overriding chart values with a local file
thelma render -e alpha -a cromwell --values-file=path/to/my-values.yaml

# Render all manifests to a directory called my-manifests
thelma render --output-dir=/tmp/my-manifests

# Render ArgoCD manifests for all Terra services in all environments
thelma render --argocd

# Render ArgoCD manifests for the Cromwell service in the alpha environment
thelma render -e alpha -a cromwell --argocd
`

// defaultOutputDir name of default output directory
const defaultOutputDir = "output"

// defaultChartSourceDir name of default chart source directory
const defaultChartSourceDir = "charts"

// renderCommand contains state and configuration for executing a render from the command-line
type renderCommand struct {
	helmfileArgs  *helmfile.Args
	renderOptions *render.Options
	flagVals      *flagValues
}

// flagNames the names of all `render`'s CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	env             string
	cluster         string
	app             string
	release         string
	chartDir        string
	chartVersion    string
	appVersion      string
	valuesFile      string
	argocd          string
	outputDir       string
	stdout          string
	parallelWorkers string
	mode            string
}{
	env:             "env",
	cluster:         "cluster",
	app:             "app",
	release:         "release",
	chartDir:        "chart-dir",
	chartVersion:    "chart-version",
	appVersion:      "app-version",
	valuesFile:      "values-file",
	argocd:          "argocd",
	outputDir:       "output-dir",
	stdout:          "stdout",
	parallelWorkers: "parallel-workers",
	mode:            "mode",
}

// flagValues is a struct for capturing flag values that are parsed by Cobra.
type flagValues struct {
	env             string
	cluster         string
	app             string
	release         string
	chartDir        string
	chartVersion    string
	appVersion      string
	valuesFile      []string
	argocd          bool
	outputDir       string
	stdout          bool
	parallelWorkers int
	mode            string
}

// NewRenderCommand constructs a new renderCommand
func NewRenderCommand() cli.ThelmaCommand {
	flagVals := &flagValues{}
	helmfileArgs := &helmfile.Args{}
	renderOptions := &render.Options{}

	cmd := &renderCommand{
		renderOptions: renderOptions,
		helmfileArgs:  helmfileArgs,
		flagVals:      flagVals,
	}

	return cmd
}

func (cmd *renderCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "render [options]"
	cobraCommand.Short = "Renders Terra Kubernetes manifests"
	cobraCommand.Long = helpMessage
	cobraCommand.SilenceUsage = true // Only print out usage error when user supplies -h/--help

	cobraCommand.Flags().StringVarP(&cmd.flagVals.env, flagNames.env, "e", "ENV", "Render manifests for a specific Terra environment only")
	cobraCommand.Flags().StringVarP(&cmd.flagVals.cluster, flagNames.cluster, "c", "CLUSTER", "Render manifests for a specific Terra cluster only")
	cobraCommand.Flags().StringVarP(&cmd.flagVals.release, flagNames.release, "r", "RELEASE", "Render manifests for a specific release only")
	cobraCommand.Flags().StringVarP(&cmd.flagVals.app, flagNames.app, "a", "APP", "Render manifests for a specific app only. (Alias for -r/--release)")
	cobraCommand.Flags().StringVar(&cmd.flagVals.chartVersion, flagNames.chartVersion, "VERSION", "Override chart version")
	cobraCommand.Flags().StringVar(&cmd.flagVals.chartDir, flagNames.chartDir, "path/to/charts", "Render from local chart directory instead of official release")
	cobraCommand.Flags().StringVar(&cmd.flagVals.appVersion, flagNames.appVersion, "VERSION", "Override application version")
	cobraCommand.Flags().StringSliceVar(&cmd.flagVals.valuesFile, flagNames.valuesFile, []string{}, "path to chart values file. Can be specified multiple times with ascending precedence (last wins)")
	cobraCommand.Flags().BoolVar(&cmd.flagVals.argocd, flagNames.argocd, false, "Render ArgoCD manifests instead of application manifests")
	cobraCommand.Flags().StringVarP(&cmd.flagVals.outputDir, flagNames.outputDir, "d", "path/to/output/dir", "Render manifests to custom output directory")
	cobraCommand.Flags().BoolVar(&cmd.flagVals.stdout, flagNames.stdout, false, "Render manifests to stdout instead of output directory")
	cobraCommand.Flags().IntVar(&cmd.flagVals.parallelWorkers, flagNames.parallelWorkers, 1, "Number of parallel workers to launch when rendering")
	cobraCommand.Flags().StringVar(&cmd.flagVals.mode, flagNames.mode, "development", `Either "development" (render from chart source directory) or "deploy" (render using released chart versions). Defaults to "development"`)
}

// PreRun argument validation and processing
func (cmd *renderCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	if len(ctx.Args()) != 0 {
		return fmt.Errorf("expected no positional arguments, got %v", ctx.Args())
	}
	flags := ctx.CobraCommand().Flags()

	if err := cmd.handleFlagAliases(flags); err != nil {
		return err
	}
	if err := cmd.checkIncompatibleFlags(flags); err != nil {
		return err
	}
	if err := cmd.fillRenderOptions(app.Config().Home(), ctx.CobraCommand().Flags()); err != nil {
		return err
	}
	if err := cmd.fillHelmfileArgs(flags); err != nil {
		return err
	}

	return nil
}

func (cmd *renderCommand) Run(app app.ThelmaApp, _ cli.RunContext) error {
	return render.DoRender(app, cmd.renderOptions, cmd.helmfileArgs)
}

func (cmd *renderCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

// fillRenderOptions populates an empty render.Options struct in accordance with user-supplied CLI options
func (cmd *renderCommand) fillRenderOptions(thelmaHome string, flags *pflag.FlagSet) error {
	flagVals := cmd.flagVals
	renderOptions := cmd.renderOptions

	// env
	if flags.Changed(flagNames.env) {
		renderOptions.Env = &flagVals.env
	}

	// cluster
	if flags.Changed(flagNames.cluster) {
		renderOptions.Cluster = &flagVals.cluster
	}

	// release name
	if flags.Changed(flagNames.release) {
		renderOptions.Release = &flagVals.release
	}

	// output dir
	if flags.Changed(flagNames.outputDir) {
		dir, err := filepath.Abs(flagVals.outputDir)
		if err != nil {
			return err
		}
		renderOptions.OutputDir = dir
	} else {
		renderOptions.OutputDir = path.Join(thelmaHome, defaultOutputDir)
		log.Debug().Msgf("Using default output dir %s", renderOptions.OutputDir)
	}

	// stdout
	renderOptions.Stdout = flagVals.stdout

	// parallelWorkers
	renderOptions.ParallelWorkers = flagVals.parallelWorkers

	// chart dir
	if flags.Changed(flagNames.chartDir) {
		chartSourceDir, err := utils.ExpandAndVerifyExists(flagVals.chartDir, "chart dir")
		if err != nil {
			return err
		}
		renderOptions.ChartSourceDir = chartSourceDir
	} else {
		renderOptions.ChartSourceDir = path.Join(thelmaHome, defaultChartSourceDir)
		log.Debug().Msgf("Using default chart source dir %s", renderOptions.ChartSourceDir)
	}

	// resolve mode
	switch flagVals.mode {
	case "development":
		renderOptions.ResolverMode = resolver.Development
	case "deploy":
		renderOptions.ResolverMode = resolver.Deploy
	default:
		return fmt.Errorf(`invalid value for --%s (must be "development" or "deploy"): %v`, flagNames.mode, flagVals.mode)
	}

	return nil
}

// fillHelmfileArgs populates an empty helfile.Args struct in accordance with user-supplied CLI options
func (cmd *renderCommand) fillHelmfileArgs(flags *pflag.FlagSet) error {
	flagVals := cmd.flagVals
	helmfileArgs := cmd.helmfileArgs

	// chart version
	if flags.Changed(flagNames.chartVersion) {
		helmfileArgs.ChartVersion = &flagVals.chartVersion
	}

	// app version
	if flags.Changed(flagNames.appVersion) {
		helmfileArgs.AppVersion = &flagVals.appVersion
	}

	// values file
	if flags.Changed(flagNames.valuesFile) {
		for _, file := range flagVals.valuesFile {
			fullpath, err := utils.ExpandAndVerifyExists(file, "values file")
			if err != nil {
				return err
			}
			helmfileArgs.ValuesFiles = append(helmfileArgs.ValuesFiles, fullpath)
		}
	}

	// argocd mode
	helmfileArgs.ArgocdMode = flagVals.argocd

	return nil
}

// given a flagset, look for legacy aliases and update the new flag.
func (cmd *renderCommand) handleFlagAliases(flags *pflag.FlagSet) error {
	// --app is a legacy alias for --release, so copy the user-supplied value over
	if flags.Changed(flagNames.app) {
		if flags.Changed(flagNames.release) {
			return fmt.Errorf("-a is a legacy alias for -r, please specify one or the other but not both")
		}
		_app := flags.Lookup(flagNames.app).Value.String()
		err := flags.Set(flagNames.release, _app)
		if err != nil {
			return fmt.Errorf("error setting --%s to --%s argument %q: %v", flagNames.release, flagNames.app, _app, err)
		}
	}

	return nil
}

func (cmd *renderCommand) checkIncompatibleFlags(flags *pflag.FlagSet) error {
	if flags.Changed(flagNames.env) && flags.Changed(flagNames.cluster) {
		return fmt.Errorf("only one of --%s or --%s may be specified", flagNames.env, flagNames.cluster)
	}

	if flags.Changed(flagNames.chartDir) {
		if flags.Changed(flagNames.chartVersion) {
			// Chart dir points at a local chart copy, chart version specifies which version to use, we can only
			// use one or the other
			return fmt.Errorf("only one of --%s or --%s may be specified", flagNames.chartDir, flagNames.chartVersion)
		}

		if !flags.Changed(flagNames.release) {
			return fmt.Errorf("--%s requires a release be specified with --%s", flagNames.chartDir, flagNames.release)
		}
	}

	if flags.Changed(flagNames.chartVersion) && !flags.Changed(flagNames.release) {
		return fmt.Errorf("--%s requires a release be specified with --%s", flagNames.chartVersion, flagNames.release)
	}

	if flags.Changed(flagNames.appVersion) {
		if !flags.Changed(flagNames.release) {
			return fmt.Errorf("--%s requires a release be specified with --%s", flagNames.appVersion, flagNames.release)
		}
		if flags.Changed(flagNames.cluster) {
			return fmt.Errorf("--%s cannot be used for cluster releases", flagNames.appVersion)
		}
	}

	if flags.Changed(flagNames.valuesFile) && !flags.Changed(flagNames.release) {
		return fmt.Errorf("--%s requires a release be specified with --%s", flagNames.valuesFile, flagNames.release)
	}

	if flags.Changed(flagNames.argocd) {
		if flags.Changed(flagNames.chartDir) || flags.Changed(flagNames.chartVersion) || flags.Changed(flagNames.appVersion) || flags.Changed(flagNames.valuesFile) {
			return fmt.Errorf("--%s cannot be used with --%s, --%s, --%s, or --%s", flagNames.argocd, flagNames.chartDir, flagNames.chartVersion, flagNames.appVersion, flagNames.valuesFile)
		}
	}

	if flags.Changed(flagNames.stdout) && flags.Changed(flagNames.outputDir) {
		// can't render to both stdout and directory
		return fmt.Errorf("--%s cannot be used with --%s", flagNames.stdout, flagNames.outputDir)
	}

	if flags.Changed(flagNames.parallelWorkers) && flags.Changed(flagNames.stdout) {
		// --parallel-workers renders manifests in parallel. For now we only support it for directory renders
		return fmt.Errorf("--%s cannot be used with --%s", flagNames.parallelWorkers, flagNames.stdout)
	}
	return nil
}

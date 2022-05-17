package render

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/selector"
	"github.com/broadinstitute/thelma/internal/thelma/render"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
	"github.com/broadinstitute/thelma/internal/thelma/render/resolver"
	"github.com/broadinstitute/thelma/internal/thelma/render/scope"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"path"
	"path/filepath"
)

const helpMessage = `Renders Terra Kubernetes manifests

Examples:

# Render manifests for leonardo
thelma render leonardo

# Render all manifests for all Terra services in all environments
thelma render ALL

# Render manifests for all Terra services in the dev environment
thelma render -e dev ALL

# Render manifests for the cromwell service in the alpha environment
thelma render -e alpha -r cromwell

# Render manifests, overriding chart values with a local file
thelma render cromwell --values-file=path/to/my-values.yaml

# Render leonardo manifests to a directory other than $THELMA_HOME/output
thelma render leonardo  --output-dir=/tmp/my-manifests
`

// defaultOutputDir name of default output directory
const defaultOutputDir = "output"

// defaultChartSourceDir name of default chart source directory
const defaultChartSourceDir = "charts"

// renderCommand contains state and configuration for executing a render from the command-line
type renderCommand struct {
	helmfileArgs  *helmfile.Args
	renderOptions *render.Options
	selector      *selector.Selector
	flagVals      *flagValues
}

// flagNames the names of all `render`'s CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	chartVersion    string
	appVersion      string
	valuesFile      string
	argocd          string
	outputDir       string
	stdout          string
	debug           string
	parallelWorkers string
	mode            string
	apps            string
	chartDir        string
	scope           string
}{
	argocd:          "argocd",
	chartDir:        "chart-dir",
	chartVersion:    "chart-version",
	appVersion:      "app-version",
	valuesFile:      "values-file",
	outputDir:       "output-dir",
	stdout:          "stdout",
	debug:           "debug",
	parallelWorkers: "parallel-workers",
	mode:            "mode",
	apps:            "apps",
	scope:           "scope",
}

// flagValues is a struct for capturing flag values that are parsed by Cobra.
type flagValues struct {
	argocd          bool
	chartVersion    string
	appVersion      string
	valuesFile      []string
	outputDir       string
	stdout          bool
	debug           bool
	parallelWorkers int
	mode            string
	apps            string
	chartDir        string
	scope           string
}

// NewRenderCommand constructs a new renderCommand
func NewRenderCommand() cli.ThelmaCommand {
	flagVals := &flagValues{}
	helmfileArgs := &helmfile.Args{}
	renderOptions := &render.Options{}

	cmd := &renderCommand{
		renderOptions: renderOptions,
		helmfileArgs:  helmfileArgs,
		selector:      selector.NewSelector(),
		flagVals:      flagVals,
	}

	return cmd
}

func (cmd *renderCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "render [options] [SERVICE]"
	cobraCommand.Short = "Renders Terra Kubernetes manifests"
	cobraCommand.Long = helpMessage

	// Release selector flags -- these flags determine which releases will be rendered
	cmd.selector.AddFlags(cobraCommand)

	// Modal flags -- these affect render behavior and can apply to both multiple and single-chart renders
	cobraCommand.Flags().BoolVar(&cmd.flagVals.argocd, flagNames.argocd, false, "Render ArgoCD manifests instead of application manifests")
	cobraCommand.Flags().StringVarP(&cmd.flagVals.outputDir, flagNames.outputDir, "d", "path/to/output/dir", "Render manifests to custom output directory")
	cobraCommand.Flags().BoolVar(&cmd.flagVals.stdout, flagNames.stdout, false, "Render manifests to stdout instead of output directory")
	cobraCommand.Flags().BoolVar(&cmd.flagVals.debug, flagNames.debug, false, "Pass --debug to helmfile to render out invalid YAML for debugging")
	cobraCommand.Flags().IntVar(&cmd.flagVals.parallelWorkers, flagNames.parallelWorkers, 1, "Number of parallel workers to launch when rendering")
	cobraCommand.Flags().StringVar(&cmd.flagVals.mode, flagNames.mode, "development", `Either "development" (render from chart source directory) or "deploy" (render using released chart versions). Defaults to "development"`)
	cobraCommand.Flags().StringVar(&cmd.flagVals.scope, flagNames.scope, "all", `One of "release" (release-scoped resources only), "destination" (environment-/cluster-wide resources, such as Argo project, only), or "all" (include both types)`)

	// Single-chart flags -- these can only be used for renders of a single chart
	cobraCommand.Flags().StringVar(&cmd.flagVals.chartVersion, flagNames.chartVersion, "", "Override chart version")
	cobraCommand.Flags().StringVar(&cmd.flagVals.appVersion, flagNames.appVersion, "", "Override application version")
	cobraCommand.Flags().StringSliceVar(&cmd.flagVals.valuesFile, flagNames.valuesFile, []string{}, "path to chart values file. Can be specified multiple times with ascending precedence (last wins)")

	// Deprecated flags
	cobraCommand.Flags().StringVarP(&cmd.flagVals.apps, flagNames.apps, "a", "", "DEPRECATED Alias for -r / --releases")
	cobraCommand.Flags().StringVar(&cmd.flagVals.chartDir, flagNames.chartDir, "", "DEPRECATED Render chart from directory other than $THELMA_HOME/charts")
}

// PreRun argument validation and processing
func (cmd *renderCommand) PreRun(app app.ThelmaApp, ctx cli.RunContext) error {
	if len(ctx.Args()) > 1 {
		return fmt.Errorf("at most 1 positional arg is allowed, got %v", ctx.Args())
	}
	flags := ctx.CobraCommand().Flags()

	if err := cmd.handleFlagAliases(flags); err != nil {
		return err
	}

	selection, err := cmd.getSelectedReleases(app, ctx.CobraCommand().Flags(), ctx.Args())
	if err != nil {
		return err
	}

	if err := cmd.fillRenderOptions(selection, app, ctx.CobraCommand().Flags()); err != nil {
		return err
	}

	if err := cmd.checkIncompatibleFlags(flags, selection); err != nil {
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

func (cmd *renderCommand) getSelectedReleases(app app.ThelmaApp, flags *pflag.FlagSet, args []string) (*selector.Selection, error) {
	state, err := app.State()
	if err != nil {
		return nil, err
	}

	// release selection
	return cmd.selector.GetSelection(state, flags, args)
}

// fillRenderOptions populates an empty render.Options struct in accordance with user-supplied CLI options
func (cmd *renderCommand) fillRenderOptions(selection *selector.Selection, app app.ThelmaApp, flags *pflag.FlagSet) error {
	flagVals := cmd.flagVals
	renderOptions := cmd.renderOptions

	if len(selection.Releases) == 0 {
		return fmt.Errorf("0 releases matched command-line arguments, nothing to render")
	}
	renderOptions.Releases = selection.Releases

	_scope, err := scope.FromString(flagVals.scope)
	if err != nil {
		return fmt.Errorf("--%s: invalid scope: %q", flagNames.scope, flagVals.scope)
	}
	if selection.IsReleaseScoped {
		if flags.Changed(flagNames.scope) && _scope != scope.Release {
			return fmt.Errorf("--%s %q cannot be used when a release is specified", flagNames.scope, flagVals.scope)
		}
		_scope = scope.Release
	}
	renderOptions.Scope = _scope

	// output dir
	if flags.Changed(flagNames.outputDir) {
		if flags.Changed(flagNames.stdout) {
			return fmt.Errorf("--%s cannot be used with --%s", flagNames.stdout, flagNames.outputDir)
		}
		dir, err := filepath.Abs(flagVals.outputDir)
		if err != nil {
			return err
		}
		renderOptions.OutputDir = dir
	} else {
		renderOptions.OutputDir = path.Join(app.Config().Home(), defaultOutputDir)
		log.Debug().Msgf("Using default output dir %s", renderOptions.OutputDir)
	}

	// stdout
	renderOptions.Stdout = flagVals.stdout

	// debug mode
	renderOptions.DebugMode = flagVals.debug

	// parallelWorkers
	if flags.Changed(flagNames.parallelWorkers) && flags.Changed(flagNames.stdout) {
		// --parallel-workers renders manifests in parallel. For now we only support it for directory renders
		return fmt.Errorf("--%s cannot be used with --%s", flagNames.parallelWorkers, flagNames.stdout)
	}
	renderOptions.ParallelWorkers = flagVals.parallelWorkers

	// chart dir
	if flags.Changed(flagNames.chartDir) {
		chartSourceDir, err := utils.ExpandAndVerifyExists(flagVals.chartDir, "chart dir")
		if err != nil {
			return err
		}
		renderOptions.ChartSourceDir = chartSourceDir
	} else {
		renderOptions.ChartSourceDir = path.Join(app.Config().Home(), defaultChartSourceDir)
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
	// --apps is a legacy alias for --releases, so copy the user-supplied value over
	if flags.Changed(flagNames.apps) {
		if flags.Changed(selector.ReleasesFlagName) {
			return fmt.Errorf("--%s is a legacy alias for --%s, please specify one or the other but not both", flagNames.apps, selector.ReleasesFlagName)
		}
		_app := flags.Lookup(flagNames.apps).Value.String()
		err := flags.Set(selector.ReleasesFlagName, _app)
		if err != nil {
			return fmt.Errorf("error setting --%s to --%s argument %q: %v", selector.ReleasesFlagName, flagNames.apps, _app, err)
		}
	}

	return nil
}

func (cmd *renderCommand) checkIncompatibleFlags(flags *pflag.FlagSet, selection *selector.Selection) error {
	if flags.Changed(flagNames.chartDir) {
		if flags.Changed(flagNames.chartVersion) {
			// Chart dir points at a local chart copy, chart version specifies which version to use, we can only
			// use one or the other
			return fmt.Errorf("only one of --%s or --%s may be specified", flagNames.chartDir, flagNames.chartVersion)
		}
	}

	if !selection.SingleChart {
		if flags.Changed(flagNames.chartVersion) || flags.Changed(flagNames.appVersion) || flags.Changed(flagNames.valuesFile) {
			return fmt.Errorf("--%s, --%s, and --%s cannot be used with selectors that match multiple charts", flagNames.chartVersion, flagNames.appVersion, flagNames.valuesFile)
		}
	}

	if !selection.AppReleasesOnly {
		if flags.Changed(flagNames.appVersion) {
			return fmt.Errorf("--%s cannot be used for cluster releases", flagNames.appVersion)
		}
	}

	if flags.Changed(flagNames.argocd) {
		if flags.Changed(flagNames.chartDir) || flags.Changed(flagNames.chartVersion) || flags.Changed(flagNames.appVersion) || flags.Changed(flagNames.valuesFile) {
			return fmt.Errorf("--%s cannot be used with --%s, --%s, --%s, or --%s", flagNames.argocd, flagNames.chartDir, flagNames.chartVersion, flagNames.appVersion, flagNames.valuesFile)
		}
	}

	return nil
}

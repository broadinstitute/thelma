package helmfile

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile/stateval"
	"github.com/broadinstitute/thelma/internal/thelma/render/resolver"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"path/filepath"
)

const cmdLogLevel = zerolog.DebugLevel
const stateValuesFilename = "stateValues.yaml"
const valuesFilename = "values.yaml"

// Args arguments for a helmfile render
type Args struct {
	ChartVersion *string  // ChartVersion optionally override chart version
	AppVersion   *string  // AppVersion optionally override application version (container image)
	ValuesFiles  []string // ValuesFiles optional list of files for overriding chart values
	ArgocdMode   bool     // ArgocdMode if true, render ArgoCD manifests instead of application manifests
}

// Options constructor arguments for a ConfigRepo
type Options struct {
	ThelmaHome       string        // ThelmaHome to terra-helmfile repo clone
	ChartSourceDir   string        // ChartSourceDir path on filesystem containing chart sources
	ChartCacheDir    string        // ChartCacheDir path to shared chart cache directory that can be re-used across renders
	ResolverMode     resolver.Mode // Resolver mode
	HelmfileLogLevel string        // HelmfileLogLevel is the --log-level argument to pass to `helmfile` command
	Stdout           bool          // Stdout if true, render to stdout instead of output directory
	OutputDir        string        // OutputDir directory where manifests should be rendered
	DebugMode        bool          // DebugMode if true, pass the --debug flag to helmfile to render out invalid yaml
	ScratchDir       string        // Scratch directory where temporary files should be written
	ShellRunner      shell.Runner  // ShellRunner shell Runner to use for executing helmfile commands
}

// ConfigRepo can be used to run `helmfile render` commands on a clone of the terra-helmfile repo
type ConfigRepo struct {
	thelmaHome       string
	chartResolver    resolver.Resolver
	helmfileLogLevel string
	stdout           bool
	debugMode        bool
	outputDir        string
	scratchDir       string
	shellRunner      shell.Runner
}

// NewConfigRepo constructs a new ConfigRepo object
func NewConfigRepo(options Options) *ConfigRepo {
	chartResolver := resolver.NewResolver(options.ShellRunner, resolver.Options{
		Mode:       options.ResolverMode,
		CacheDir:   options.ChartCacheDir,
		SourceDir:  options.ChartSourceDir,
		ScratchDir: path.Join(options.ScratchDir, "resolver"),
	})

	return &ConfigRepo{
		thelmaHome:       options.ThelmaHome,
		chartResolver:    chartResolver,
		helmfileLogLevel: options.HelmfileLogLevel,
		stdout:           options.Stdout,
		debugMode:        options.DebugMode,
		outputDir:        options.OutputDir,
		scratchDir:       path.Join(options.ScratchDir, "helmfile"),
		shellRunner:      options.ShellRunner,
	}
}

// CleanOutputDirectoryIfEnabled cleans the output directory before rendering
func (r *ConfigRepo) CleanOutputDirectoryIfEnabled() error {
	if r.stdout {
		// No need to clean output directory if we're rendering to stdout
		return nil
	}

	if _, err := os.Stat(r.outputDir); os.IsNotExist(err) {
		// Output dir does not exist, nothing to clean up
		return nil
	}

	// Delete any files that exist inside the output directory.
	log.Debug().Msgf("Cleaning output directory: %s", r.outputDir)

	// This code would be simpler if we could just call os.RemoveAll() on the
	// output directory itself, but in some cases the output directory is a volume
	// mount in a Docker container, and trying to remove it throws an error.
	// So we remove all its contents instead.
	dir, err := os.ReadDir(r.outputDir)
	if err != nil {
		return err
	}

	for _, file := range dir {
		filePath := path.Join([]string{r.outputDir, file.Name()}...)
		log.Debug().Msgf("Deleting %s", filePath)

		err = os.RemoveAll(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// HelmUpdate updates Helm repo indexes.
func (r *ConfigRepo) HelmUpdate() error {
	log.Info().Msg("Updating Helm repo indexes")

	var args []string
	if r.helmfileLogLevel != "" {
		args = append(args, fmt.Sprintf("--log-level=%s", r.helmfileLogLevel))
	}
	args = append(args, "--allow-no-matching-release", "repos")

	cmd := shell.Command{
		Prog: ProgName,
		Args: args,
		Dir:  r.thelmaHome,
	}

	return r.runCmd(cmd)
}

func (r *ConfigRepo) RenderForDestination(destination terra.Destination, args *Args) error {
	if args.ArgocdMode {
		if len(destination.Releases()) == 0 {
			log.Debug().Msgf("%s %s has no releases, won't render ArgoCD project", destination.Type(), destination.Name())
		} else {
			return r.renderArgocdProjectManifests(destination)
		}
	}
	return nil
}

func (r *ConfigRepo) RenderForRelease(release terra.Release, args *Args) error {
	if args.ArgocdMode {
		return r.renderArgocdApplicationManifests(release)
	} else {
		return r.renderApplicationManifests(release, args)
	}
}

// Render Argo project manifests for the given destination
func (r *ConfigRepo) renderArgocdProjectManifests(destination terra.Destination) error {
	outputDir := path.Join(r.outputDir, destination.Name(), "terra-argocd-project")

	// Generate state values file
	stateValues := stateval.BuildArgoProjectValues(destination)
	stateValuesFile := r.scratchPath(destination.Name(), "argo", "project", stateValuesFilename)
	if err := writeTemporaryValuesFile(stateValues, stateValuesFile); err != nil {
		return fmt.Errorf("error generating argo project state values file for %s %s: %v", destination.Type(), destination.Name(), err)
	}

	// Generate chart values file with destinations
	values := argocd.GetDestinationValues(destination)
	valuesFile := r.scratchPath(destination.Name(), "argo", "project", valuesFilename)
	if err := writeTemporaryValuesFile(values, valuesFile); err != nil {
		return fmt.Errorf("error generating argo project values file for %s %s: %v", destination.Type(), destination.Name(), err)
	}

	cmd := newCmd()
	cmd.setStateValuesFile(stateValuesFile)
	cmd.setOutputDir(outputDir)
	cmd.setStdout(r.stdout)
	cmd.setDebugMode(r.debugMode)
	cmd.setDir(path.Join(r.thelmaHome, "argocd", "project"))
	cmd.setLogLevel(r.helmfileLogLevel)
	cmd.setSkipDeps(true) // argocd project chart is local & has no dependencies
	cmd.addValuesFiles(valuesFile)

	log.Info().Msgf("Rendering ArgoCD manifests for %s %s", destination.Name(), destination.Type())

	return r.runHelmfile(cmd)
}

// Render Argo manifests for the given release
func (r *ConfigRepo) renderArgocdApplicationManifests(release terra.Release) error {
	stateValues := stateval.BuildArgoAppValues(release)
	stateValuesFile := r.scratchPath(release.Destination().Name(), release.Name(), "argocd", "application", stateValuesFilename)
	if err := writeTemporaryValuesFile(stateValues, stateValuesFile); err != nil {
		return fmt.Errorf("error rendering argo app state values for release %s in %s %s: %v", release.Name(), release.Destination().Type(), release.Destination().Name(), err)
	}

	dir := fmt.Sprintf("terra-argocd-app-%s", release.Name())
	outputDir := path.Join(r.outputDir, release.Destination().Name(), dir)

	cmd := newCmd()
	cmd.setStateValuesFile(stateValuesFile)
	cmd.setOutputDir(outputDir)
	cmd.setStdout(r.stdout)
	cmd.setDebugMode(r.debugMode)
	cmd.setDir(path.Join(r.thelmaHome, "argocd", "application"))
	cmd.setLogLevel(r.helmfileLogLevel)
	cmd.setSkipDeps(true) // argocd application chart is local & has no dependencies

	log.Info().Msgf("Rendering ArgoCD manifests for %s in %s", release.Name(), release.Destination().Name())

	return r.runHelmfile(cmd)
}

// Render application manifests for the given release
func (r *ConfigRepo) renderApplicationManifests(release terra.Release, args *Args) error {
	chartVersion := release.ChartVersion()
	if args.ChartVersion != nil {
		log.Debug().Msgf("Overriding default chart version %s for %s with %s", chartVersion, release.Name(), *args.ChartVersion)
		chartVersion = *args.ChartVersion
	}

	resolvedChart, err := r.chartResolver.Resolve(resolver.ChartRelease{
		Name:    release.ChartName(),
		Repo:    release.Repo(),
		Version: chartVersion,
	})
	if err != nil {
		return fmt.Errorf("error resolving chart for release %s in %s %s: %v", release.Name(), release.Destination().Type(), release.Destination().Name(), err)
	}

	outputDir := path.Join(r.outputDir, release.Destination().Name(), release.Name())

	stateValues := stateval.BuildAppValues(release, resolvedChart.Path())
	stateValues = overrideAppVersionIfNeeded(release, args, stateValues)

	stateValuesFile := r.scratchPath(release.Destination().Name(), release.Name(), stateValuesFilename)
	if err = writeTemporaryValuesFile(stateValues, stateValuesFile); err != nil {
		return fmt.Errorf("error rendering state values for release %s in %s %s: %v", release.Name(), release.Destination().Type(), release.Destination().Name(), err)
	}

	cmd := newCmd()
	cmd.setStateValuesFile(stateValuesFile)
	cmd.setOutputDir(outputDir)
	cmd.setStdout(r.stdout)
	cmd.setDebugMode(r.debugMode)
	cmd.setDir(r.thelmaHome)
	cmd.setLogLevel(r.helmfileLogLevel)

	// resolver runs `helm dependency update` on local charts, so we always set --skip-deps to save time
	cmd.setSkipDeps(true)
	// tests are noisy so exclude them
	cmd.setSkipTests(true)

	cmd.addValuesFiles(args.ValuesFiles...)

	logEvent := log.Info().
		Str("chartVersion", resolvedChart.Version()).
		Str("chartSource", resolvedChart.SourceDescription())
	if release.IsAppRelease() {
		logEvent.Str("appVersion", stateValues.Release.AppVersion)
	}
	logEvent.Msgf("Rendering %s in %s", release.Name(), release.Destination().Name())

	return r.runHelmfile(cmd)
}

func (r *ConfigRepo) runHelmfile(cmd *Cmd) error {
	err := r.runCmd(cmd.toShellCommand())
	if err != nil {
		return err
	}

	if !r.stdout {
		return normalizeOutputDir(cmd.outputDir)
	}

	return nil
}

func (r *ConfigRepo) runCmd(cmd shell.Command) error {
	level := cmdLogLevel

	var stdoutWriter io.Writer
	if r.stdout {
		stdoutWriter = os.Stdout
	}
	return r.shellRunner.Run(cmd, func(opts *shell.RunOptions) {
		opts.LogLevel = level
		opts.Stdout = stdoutWriter
	})
}

// Normalize output directories so they match what was produced by earlier iterations of the render tool.
//
// Old behavior
// ------------
// For regular releases:
//  output/dev/helmfile-b47efc70-leonardo/leonardo
//  ->
//  output/dev/leonardo/leonardo
//
// For ArgoCD:
//  output/dev/helmfile-b47efc70-terra-argocd-app-leonardo/terra-argocd-app
//  ->
//  output/dev/terra-argocd-app-leonardo/terra-argocd-app
//
//  output/dev/helmfile-b47efc70-terra-argocd-project/terra-argocd-project
//  ->
//  output/dev/terra-argocd-project/terra-argocd-project
//
// New behavior
// ------------
// For regular releases:
//  output/dev/leonardo/helmfile-b47efc70-leonardo/leonardo
//  ->
//  output/dev/leonardo/leonardo
//
// For ArgoCD:
//  output/dev/terra-argocd-app-leonardo/helmfile-b47efc70-terra-argocd-app-leonardo/terra-argocd-app
//  ->
//  output/dev/terra-argocd-app-leonardo/terra-argocd-app
//
//  output/dev/terra-argocd-project/helmfile-b47efc70-terra-argocd-project/terra-argocd-project
//  ->
//  output/dev/terra-argocd-project/terra-argocd-project
//
// normalizeOutputDir removes "helmfile-.*" directories from helmfile output paths.
// this makes it possible to easily run diff -r on render outputs from different branches
func normalizeOutputDir(outputDir string) error {
	glob := path.Join(outputDir, "helmfile-*", "*")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("error globbing rendered templates %s: %v", glob, err)
	}

	if len(matches) != 1 {
		return fmt.Errorf("expected exactly one match for %s, got %d: %v", glob, len(matches), matches)
	}

	match := matches[0]
	dest := path.Join(path.Dir(path.Dir(match)), path.Base(match))
	log.Debug().Msgf("Renaming %s to %s", match, dest)

	if err := os.Rename(match, dest); err != nil {
		return err
	}
	if err := os.Remove(path.Dir(match)); err != nil {
		return err
	}

	return nil
}

// Like path.Join, but prefixes with the scratch directory path
func (r *ConfigRepo) scratchPath(pathComponents ...string) string {
	return path.Join(r.scratchDir, path.Join(pathComponents...))
}

// Marshal structured data to YAML and write to the given file
func writeTemporaryValuesFile(values interface{}, filename string) error {
	if err := os.MkdirAll(path.Dir(filename), 0775); err != nil {
		return fmt.Errorf("error creating parent directories for temporary values file %s: %v", filename, err)

	}
	output, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("error marshaling values for %s to YAML: %v (content: %v)", filename, err, values)
	}

	if err := os.WriteFile(filename, output, 0666); err != nil {
		return fmt.Errorf("error writing temporary values file %s: %v", filename, err)
	}

	return nil
}

// Override app version in state values if it was set on the command line with --app-version
func overrideAppVersionIfNeeded(release terra.Release, args *Args, stateValues stateval.AppValues) stateval.AppValues {
	if release.Type() == terra.AppReleaseType {
		if args.AppVersion != nil {
			originalVersion := stateValues.Release.AppVersion
			log.Debug().Msgf("Overriding default app version %s for release %s with %s", originalVersion, release.Name(), *args.AppVersion)
			stateValues.Release.AppVersion = *args.AppVersion
		}
	} else if args.AppVersion != nil {
		log.Warn().Msgf("Ignoring --app-version %s; --app-version is not supported for cluster releases", *args.AppVersion)
	}

	return stateValues
}

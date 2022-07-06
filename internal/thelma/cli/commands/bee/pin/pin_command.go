package pin

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/builders"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const helpMessage = `Override the version of a service that is deployed to a BEE.

Examples:

# Pin leonardo application image to tag v100
thelma bee pin -n swat-grungy-puma sam --app-version=v100

# Pin sam helm chart to version 0.10.3
thelma bee pin -n swat-grungy-puma sam --chart-version=0.10.3

# Pin sam to the terra-helmfile PR branch my-pr-1
thelma bee pin -n swat-grungy-puma sam --terra-helmfile-ref=my-pr-1

# Pin sam to the firecloud-develop PR branch my-pr-1
thelma bee pin -n swat-grungy-puma sam --firecloud-develop-ref=my-pr-1

# Pin all services in a BEE to the terra-helmfile PR branch my-pr-1
thelma bee pin -n swat-grungy-puma ALL --terra-helmfile-ref=my-pr-1

# Pin all services in a BEE to the firecloud-develop PR branch my-pr-1
thelma bee pin -n swat-grungy-puma ALL --firecloud-develop-ref=my-pr-1

# Pin all services in a BEE to versions described in the given file, with a format like:
#   {
#      "sam": {
#        "appVersion": "my-image-tag",
#        "firecloudDevelopRef": "my-image-tag",
#      },
#      "leonardo": {
#        "firecloud"
#      },
#      ...
#   }
thelma bee pin -n swat-grungy-puma sam --app-version=v100

`

type options struct {
	name                string
	appVersion          string
	chartVersion        string
	terraHelmfileRef    string
	firecloudDevelopRef string
	versionsFile        string
	versionsFormat      string
	buildNumber         int
	ifExists            bool
	sync                bool
	waitHealthy         bool
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name                string
	appVersion          string
	chartVersion        string
	terraHelmfileRef    string
	firecloudDevelopRef string
	versionsFile        string
	versionsFormat      string
	buildNumber         string
	releases            string
	ifExists            string
	sync                string
	waitHealthy         string
}{
	name:                "name",
	appVersion:          "app-version",
	chartVersion:        "chart-version",
	terraHelmfileRef:    "terra-helmfile-ref",
	firecloudDevelopRef: "firecloud-develop-ref",
	versionsFile:        "versions-file",
	versionsFormat:      "versions-format",
	buildNumber:         "build-number",
	releases:            "releases",
	ifExists:            "if-exists",
	sync:                "sync",
	waitHealthy:         "wait-healthy",
}

type pinCommand struct {
	options       options
	serviceScoped bool
	versions      map[string]terra.VersionOverride
}

func NewBeePinCommand() cli.ThelmaCommand {
	return &pinCommand{}
}

func (cmd *pinCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "pin [SERVICE] [options]"
	cobraCommand.Short = "Pin a BEE to specific version"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to pin")
	cobraCommand.Flags().StringVar(&cmd.options.appVersion, flagNames.appVersion, "", "Pin to specific image tag")
	cobraCommand.Flags().StringVar(&cmd.options.chartVersion, flagNames.chartVersion, "", "Pin to specific Helm chart version")
	cobraCommand.Flags().StringVar(&cmd.options.terraHelmfileRef, flagNames.terraHelmfileRef, "", "Pin to specific terra-helmfile ref")
	cobraCommand.Flags().StringVar(&cmd.options.firecloudDevelopRef, flagNames.firecloudDevelopRef, "", "Pin to specific firecloud-develop ref")
	cobraCommand.Flags().StringVar(&cmd.options.versionsFile, flagNames.versionsFile, "", "Path to versions file")
	cobraCommand.Flags().StringVar(&cmd.options.versionsFormat, flagNames.versionsFormat, "yaml", fmt.Sprintf("Format of --%s. One of: %s", flagNames.versionsFile, utils.QuoteJoin(versionFormats())))
	cobraCommand.Flags().IntVar(&cmd.options.buildNumber, flagNames.buildNumber, 0, "Configure environment's currently running build number (for use in CI/CD pipelines)")
	cobraCommand.Flags().BoolVar(&cmd.options.ifExists, flagNames.ifExists, false, "Do not return an error if the BEE does not exist")
	cobraCommand.Flags().BoolVar(&cmd.options.sync, flagNames.sync, true, "Sync all services in BEE after updating versions")
	cobraCommand.Flags().BoolVar(&cmd.options.waitHealthy, flagNames.waitHealthy, true, "Wait for BEE's Argo apps to become healthy after syncing")
}

func (cmd *pinCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	// validate --name
	if !flags.Changed(flagNames.name) {
		return fmt.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.options.name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return fmt.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}

	return nil
}

func (cmd *pinCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}

	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}

	if env == nil {
		if cmd.options.ifExists {
			log.Warn().Msgf("Could not pin %s, no BEE by that name exists", cmd.options.name)
			return nil
		}
		return fmt.Errorf("--%s: unknown bee %q", flagNames.name, cmd.options.name)
	}

	if err = cmd.loadVersions(ctx, env); err != nil {
		return err
	}

	if !cmd.serviceScoped && ctx.CobraCommand().Flags().Changed(flagNames.terraHelmfileRef) {
		log.Info().Msgf("Pinning environment %s to terra-helmfile ref: %s", cmd.options.name, cmd.options.terraHelmfileRef)
		if err = state.Environments().PinEnvironmentToTerraHelmfileRef(cmd.options.name, cmd.options.terraHelmfileRef); err != nil {
			return err
		}
	}

	versions, err := state.Environments().PinVersions(cmd.options.name, cmd.versions)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated version overrides for %s", cmd.options.name)

	if cmd.options.buildNumber != 0 {
		oldBuildNumber, err := state.Environments().SetBuildNumber(cmd.options.name, cmd.options.buildNumber)
		if err != nil {
			return err
		}
		log.Info().Msgf("Set build number to %d for %s (was: %d)", cmd.options.buildNumber, cmd.options.name, oldBuildNumber)
	}

	if err = bees.RefreshBeeGenerator(); err != nil {
		return err
	}
	if err = bees.SyncEnvironmentGenerator(env); err != nil {
		return err
	}

	if cmd.options.sync {
		if err = bees.SyncArgoAppsIn(env, func(options *argocd.SyncOptions) {
			options.WaitHealthy = cmd.options.waitHealthy
		}); err != nil {
			return err
		}
	}

	log.Info().Msgf("Full set of overrides for %s:", cmd.options.name)
	ctx.SetOutput(versions)
	return nil
}

func (cmd *pinCommand) loadVersions(ctx cli.RunContext, env terra.Environment) error {
	flags := ctx.CobraCommand().Flags()

	if len(ctx.Args()) == 0 {
		cmd.serviceScoped = false
		if flags.Changed(flagNames.appVersion) || flags.Changed(flagNames.chartVersion) {
			return fmt.Errorf("--%s and --%s can only be used with a positional argument", flagNames.appVersion, flagNames.chartVersion)
		}
		if flags.Changed(flagNames.versionsFile) {
			return cmd.readVersionsFromFile()
		} else if flags.Changed(flagNames.terraHelmfileRef) || flags.Changed(flagNames.firecloudDevelopRef) {
			return cmd.buildVersionsForAllServices(env)
		} else {
			return fmt.Errorf("please specify --%s or --%s/--%s", flagNames.versionsFile, flagNames.terraHelmfileRef, flagNames.firecloudDevelopRef)
		}
	} else {
		if flags.Changed(flagNames.versionsFile) {
			return fmt.Errorf("--%s cannot be used with a positional argument", flagNames.versionsFile)
		}
		cmd.serviceScoped = true
		serviceName := ctx.Args()[0]
		return cmd.buildVersionsForService(env, serviceName)
	}
}

func (cmd *pinCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

func (cmd *pinCommand) readVersionsFromFile() error {
	content, err := os.ReadFile(cmd.options.versionsFile)
	if err != nil {
		return err
	}
	versions, err := parseVersions(cmd.options.versionsFormat, content)
	if err != nil {
		return err
	}

	cmd.versions = cmd.applyGitRefOverrides(versions)
	return nil
}

func (cmd *pinCommand) applyGitRefOverrides(versions map[string]terra.VersionOverride) map[string]terra.VersionOverride {
	result := make(map[string]terra.VersionOverride)
	for releaseName, override := range versions {
		if cmd.options.terraHelmfileRef != "" {
			override.TerraHelmfileRef = cmd.options.terraHelmfileRef
		}
		if cmd.options.firecloudDevelopRef != "" {
			override.FirecloudDevelopRef = cmd.options.firecloudDevelopRef
		}
		result[releaseName] = override
	}
	return result
}

func (cmd *pinCommand) buildVersionsForAllServices(env terra.Environment) error {
	versions := make(map[string]terra.VersionOverride)
	for _, release := range env.Releases() {
		versions[release.Name()] = terra.VersionOverride{
			// app and chart version omitted because it makes no sense to use them for multiple services
			TerraHelmfileRef:    cmd.options.terraHelmfileRef,
			FirecloudDevelopRef: cmd.options.firecloudDevelopRef,
		}
	}
	cmd.versions = versions
	return nil
}

func (cmd *pinCommand) buildVersionsForService(env terra.Environment, serviceName string) error {
	var release terra.Release
	for _, r := range env.Releases() {
		if r.Name() == serviceName {
			release = r
		}
	}

	if release == nil {
		return fmt.Errorf("error setting version overrides: service %s does not exist in BEE %s", serviceName, env.Name())
	}

	versions := make(map[string]terra.VersionOverride)
	versions[release.Name()] = terra.VersionOverride{
		AppVersion:          cmd.options.appVersion,
		ChartVersion:        cmd.options.chartVersion,
		TerraHelmfileRef:    cmd.options.terraHelmfileRef,
		FirecloudDevelopRef: cmd.options.firecloudDevelopRef,
	}
	cmd.versions = versions
	return nil
}

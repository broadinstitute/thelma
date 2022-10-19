package pinflags

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/spf13/cobra"
	"os"
)

type flagValues struct {
	terraHelmfileRef    string
	firecloudDevelopRef string
	versionsFile        string
	versionsFormat      string
	fromEnv             string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	terraHelmfileRef    string
	firecloudDevelopRef string
	versionsFile        string
	versionsFormat      string
	fromEnv             string
}{
	terraHelmfileRef:    "terra-helmfile-ref",
	firecloudDevelopRef: "firecloud-develop-ref",
	versionsFile:        "versions-file",
	versionsFormat:      "versions-format",
	fromEnv:             "from-env",
}

type pinFlags struct {
	options flagValues
}

// PinFlags adds version pinning CLI flags to a cobra command and supports converting those flags to a bee.PinOptions struct
type PinFlags interface {
	// AddFlags add version pinning flags such as --versions-file, --terra-helmfile-ref, and so forth to a Cobra command
	AddFlags(*cobra.Command)
	// GetPinOptions can be called during a Run function to get a bee.PinOptions populated with settings from version pinning CLI flags
	GetPinOptions(thelmaApp app.ThelmaApp, rc cli.RunContext) (bee.PinOptions, error)
}

// NewPinFlags returns a new PinFlags
func NewPinFlags() PinFlags {
	return &pinFlags{}
}

func (p *pinFlags) AddFlags(cobraCommand *cobra.Command) {
	cobraCommand.Flags().StringVar(&p.options.terraHelmfileRef, flagNames.terraHelmfileRef, "", "Pin BEE to specific terra-helmfile branch (instead of master)")
	cobraCommand.Flags().StringVar(&p.options.firecloudDevelopRef, flagNames.firecloudDevelopRef, "", "Pin BEE to specific firecloud-develop branch (instead of dev)")
	cobraCommand.Flags().StringVar(&p.options.versionsFile, flagNames.versionsFile, "", `Path to file containing application version overrides (see "thelma bee pin --help" for more info)`)
	cobraCommand.Flags().StringVar(&p.options.versionsFormat, flagNames.versionsFormat, "yaml", fmt.Sprintf("Format of --%s. One of: %s", flagNames.versionsFile, utils.QuoteJoin(versionFormats())))
	cobraCommand.Flags().StringVar(&p.options.fromEnv, flagNames.fromEnv, "", "Name of an environment to pull versions from")
}

func (p *pinFlags) GetPinOptions(thelmaApp app.ThelmaApp, rc cli.RunContext) (bee.PinOptions, error) {
	var pinOpts bee.PinOptions

	pinOpts.Flags.TerraHelmfileRef = p.options.terraHelmfileRef
	pinOpts.Flags.FirecloudDevelopRef = p.options.firecloudDevelopRef

	fileOverrides, err := p.loadReleaseOverrides(thelmaApp, rc)
	if err != nil {
		return pinOpts, err
	}
	pinOpts.FileOverrides = fileOverrides

	return pinOpts, nil
}

func (p *pinFlags) loadReleaseOverrides(thelmaApp app.ThelmaApp, rc cli.RunContext) (map[string]terra.VersionOverride, error) {
	flags := rc.CobraCommand().Flags()
	if flags.Changed(flagNames.fromEnv) {
		if flags.Changed(flagNames.versionsFile) {
			return nil, fmt.Errorf("either %s or %s can be specified, but not both", flagNames.versionsFile, flagNames.fromEnv)
		}
		return p.loadReleaseOverridesFromEnv(thelmaApp)
	} else if flags.Changed(flagNames.versionsFile) {
		return p.loadReleaseOverridesFromFile()
	} else {
		// return empty map if no overrides env or file was supplied
		return make(map[string]terra.VersionOverride), nil
	}
}

func (p *pinFlags) loadReleaseOverridesFromFile() (map[string]terra.VersionOverride, error) {
	file := p.options.versionsFile
	format := p.options.versionsFormat

	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	versions, err := parseVersions(format, content)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (p *pinFlags) loadReleaseOverridesFromEnv(thelmaApp app.ThelmaApp) (map[string]terra.VersionOverride, error) {
	state, err := thelmaApp.State()
	if err != nil {
		return nil, err
	}
	env, err := state.Environments().Get(p.options.fromEnv)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, fmt.Errorf("--%s: no such environment %q", flagNames.fromEnv, p.options.fromEnv)
	}

	defaultTerraHelmfileRef := "master"
	defaultFirecloudDevelopRef := "dev"
	if env.Name() == "alpha" || env.Name() == "staging" || env.Name() == "prod" {
		defaultFirecloudDevelopRef = env.Name()
	}

	result := make(map[string]terra.VersionOverride)
	for _, release := range env.Releases() {
		var appVersion string
		if release.IsAppRelease() {
			appVersion = release.(terra.AppRelease).AppVersion()
		}
		override := terra.VersionOverride{
			AppVersion:          appVersion,
			ChartVersion:        release.ChartVersion(),
			TerraHelmfileRef:    release.TerraHelmfileRef(),
			FirecloudDevelopRef: release.FirecloudDevelopRef(),
		}
		if override.TerraHelmfileRef == "" {
			override.TerraHelmfileRef = defaultTerraHelmfileRef
		}
		if override.FirecloudDevelopRef == "" {
			override.FirecloudDevelopRef = defaultFirecloudDevelopRef
		}
		result[release.Name()] = override
	}
	return result, nil
}

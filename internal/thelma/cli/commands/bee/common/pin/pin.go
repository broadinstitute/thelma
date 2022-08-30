package pin

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/spf13/cobra"
	"os"
)

type options struct {
	terraHelmfileRef    string
	firecloudDevelopRef string
	versionsFile        string
	versionsFormat      string
	buildNumber         int
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	terraHelmfileRef    string
	firecloudDevelopRef string
	versionsFile        string
	versionsFormat      string
	buildNumber         string
}{
	terraHelmfileRef:    "terra-helmfile-ref",
	firecloudDevelopRef: "firecloud-develop-ref",
	versionsFile:        "versions-file",
	versionsFormat:      "versions-format",
	buildNumber:         "build-number",
}

type optionsBuilder struct {
	options options
}

// OptionsBuilder adds version pinning CLI flags to a cobra command and supports converting those flags to a bee.PinOptions struct
type OptionsBuilder interface {
	// AddFlags add version pinning flags such as --versions-file, --terra-helmfile-ref, and so forth to a Cobra command
	AddFlags(*cobra.Command)
	// LoadPinOptions can be called during a Run function to get a bee.PinOptions populated with settings from version pinning CLI flags
	LoadPinOptions(rc cli.RunContext) (bee.PinOptions, error)
}

// NewPinOptionsBuilder returns a new PinOptionsBuilder
func NewPinOptionsBuilder() OptionsBuilder {
	return &optionsBuilder{}
}

func (l *optionsBuilder) AddFlags(cobraCommand *cobra.Command) {
	cobraCommand.Flags().StringVar(&l.options.terraHelmfileRef, flagNames.terraHelmfileRef, "", "Pin BEE to specific terra-helmfile branch (instead of master)")
	cobraCommand.Flags().StringVar(&l.options.firecloudDevelopRef, flagNames.firecloudDevelopRef, "", "Pin BEE to specific firecloud-develop branch (instead of dev)")
	cobraCommand.Flags().StringVar(&l.options.versionsFile, flagNames.versionsFile, "", "Path to file containing application version overrides (see `thelma bee pin --help` for more info)")
	cobraCommand.Flags().StringVar(&l.options.versionsFormat, flagNames.versionsFormat, "yaml", fmt.Sprintf("Format of --%s. One of: %s", flagNames.versionsFile, utils.QuoteJoin(versionFormats())))
	cobraCommand.Flags().IntVar(&l.options.buildNumber, flagNames.buildNumber, 0, "Configure environment's currently running build number (for use in CI/CD pipelines)")
}

func (l *optionsBuilder) LoadPinOptions(rc cli.RunContext) (bee.PinOptions, error) {
	var overrides bee.PinOptions

	overrides.Flags.BuildNumber = l.options.buildNumber
	overrides.Flags.TerraHelmfileRef = l.options.terraHelmfileRef
	overrides.Flags.FirecloudDevelopRef = l.options.firecloudDevelopRef

	fileOverrides, err := l.loadReleaseOverridesFromFile(rc)
	if err != nil {
		return overrides, err
	}

	overrides.FileOverrides = fileOverrides
	return overrides, nil
}

func (l *optionsBuilder) loadReleaseOverridesFromFile(rc cli.RunContext) (map[string]terra.VersionOverride, error) {
	if !rc.CobraCommand().Flags().Changed(flagNames.versionsFile) {
		// return empty map if no overrides file was supplied
		return make(map[string]terra.VersionOverride), nil
	}

	file := l.options.versionsFile
	format := l.options.versionsFormat

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

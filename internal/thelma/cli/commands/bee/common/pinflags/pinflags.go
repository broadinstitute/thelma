package pinflags

import (
	"fmt"
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
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	terraHelmfileRef    string
	firecloudDevelopRef string
	versionsFile        string
	versionsFormat      string
}{
	terraHelmfileRef:    "terra-helmfile-ref",
	firecloudDevelopRef: "firecloud-develop-ref",
	versionsFile:        "versions-file",
	versionsFormat:      "versions-format",
}

type pinFlags struct {
	options flagValues
}

// PinFlags adds version pinning CLI flags to a cobra command and supports converting those flags to a bee.PinOptions struct
type PinFlags interface {
	// AddFlags add version pinning flags such as --versions-file, --terra-helmfile-ref, and so forth to a Cobra command
	AddFlags(*cobra.Command)
	// GetPinOptions can be called during a Run function to get a bee.PinOptions populated with settings from version pinning CLI flags
	GetPinOptions(rc cli.RunContext) (bee.PinOptions, error)
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
}

func (p *pinFlags) GetPinOptions(rc cli.RunContext) (bee.PinOptions, error) {
	var overrides bee.PinOptions

	overrides.Flags.TerraHelmfileRef = p.options.terraHelmfileRef
	overrides.Flags.FirecloudDevelopRef = p.options.firecloudDevelopRef

	fileOverrides, err := p.loadReleaseOverridesFromFile(rc)
	if err != nil {
		return overrides, err
	}

	overrides.FileOverrides = fileOverrides
	return overrides, nil
}

func (p *pinFlags) loadReleaseOverridesFromFile(rc cli.RunContext) (map[string]terra.VersionOverride, error) {
	if !rc.CobraCommand().Flags().Changed(flagNames.versionsFile) {
		// return empty map if no overrides file was supplied
		return make(map[string]terra.VersionOverride), nil
	}

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

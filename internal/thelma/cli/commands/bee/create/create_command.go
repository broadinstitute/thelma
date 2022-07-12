package create

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/validate"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Create a new BEE (Branch Engineering Environment) from a template

Examples:

# Create a hybrid BEE from the swatomation template
thelma bee create \
  --name=swat-grungy-puma \
  --template=swatomation \
  --hybrid \
  --fiab-name=fiab-automation-grungy-puma \
  --fiab-ip=35.36.37.38
`

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name             string
	template         string
	hybrid           string
	fiabName         string
	fiabIP           string
	generatorOnly    string
	waitHealthy      string
	terraHelmfileRef string
}{
	name:             "name",
	template:         "template",
	hybrid:           "hybrid",
	fiabName:         "fiab-name",
	fiabIP:           "fiab-ip",
	generatorOnly:    "generator-only",
	waitHealthy:      "wait-healthy",
	terraHelmfileRef: "terra-helmfile-ref",
}

type createCommand struct {
	name    string
	options bee.CreateOptions
}

func NewBeeCreateCommand() cli.ThelmaCommand {
	return &createCommand{}
}

func (cmd *createCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "create"
	cobraCommand.Short = "Create a new BEE from a template"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.name, flagNames.name, "n", "NAME", "Required. Name for this BEE")
	cobraCommand.Flags().StringVarP(&cmd.options.Template, flagNames.template, "t", "swatomation", "Template to use for this BEE")
	cobraCommand.Flags().BoolVar(&cmd.options.Hybrid, flagNames.hybrid, false, "Set to true to create a hybrid (connected to a Fiab) environment")
	cobraCommand.Flags().StringVar(&cmd.options.Fiab.Name, flagNames.fiabName, "FIAB", "Name of the Fiab this hybrid environment should be connected to")
	cobraCommand.Flags().StringVar(&cmd.options.Fiab.IP, flagNames.fiabIP, "IP", "Public IP address of the Fiab this hybrid environment should be connected to")
	cobraCommand.Flags().BoolVar(&cmd.options.SyncGeneratorOnly, flagNames.generatorOnly, false, "Sync the BEE generator but not the BEE's Argo apps")
	cobraCommand.Flags().BoolVar(&cmd.options.WaitHealthy, flagNames.waitHealthy, false, "Wait for BEE's Argo apps to become healthy after syncing")
	cobraCommand.Flags().StringVar(&cmd.options.TerraHelmfileRef, flagNames.terraHelmfileRef, "", "Custom terra-helmfile branch/ref")
}

func (cmd *createCommand) PreRun(thelmaApp app.ThelmaApp, ctx cli.RunContext) error {
	// validate --name
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("no environment name specified; --%s is required", flagNames.name)
	}
	if strings.TrimSpace(cmd.name) == "" {
		log.Warn().Msg("Is Thelma running in CI? Check that you're setting the name of your environment when running your job")
		return fmt.Errorf("no environment name specified; --%s was passed but no name was given", flagNames.name)
	}
	if err := validate.EnvironmentName(cmd.name); err != nil {
		return fmt.Errorf("--%s: %q is not a valid environment name: %v", flagNames.name, cmd.name, err)
	}

	// validate --template
	bees, err := builders.NewBees(thelmaApp)
	if err != nil {
		return err
	}
	_, err = bees.GetTemplate(cmd.options.Template)
	if err != nil {
		return fmt.Errorf("--%s: %v", flagNames.template, err)
	}

	// validate --hybrid arguments
	if cmd.options.Hybrid {
		if !ctx.CobraCommand().Flags().Changed(flagNames.fiabName) || cmd.options.Fiab.Name == "" {
			return fmt.Errorf("--%s is required for hybrid environments", flagNames.fiabName)
		}
		if !ctx.CobraCommand().Flags().Changed(flagNames.fiabIP) {
			return fmt.Errorf("--%s is required for hybrid environments", flagNames.fiabIP)
		}
		if !utils.IsIPV4Address(cmd.options.Fiab.IP) {
			return fmt.Errorf("--%s: %q is not a valid ipv4 address", flagNames.fiabIP, cmd.options.Fiab.IP)
		}
	}

	return nil
}

func (cmd *createCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}
	env, err := bees.CreateWith(cmd.name, cmd.options)
	if env != nil {
		ctx.SetOutput(views.ForTerraEnv(env))
	}
	return err
}

func (cmd *createCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

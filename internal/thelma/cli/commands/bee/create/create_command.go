package create

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
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

type options struct {
	name     string
	template string
	hybrid   bool
	fiabName string
	fiabIP   string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name     string
	template string
	hybrid   string
	fiabName string
	fiabIP   string
}{
	name:     "name",
	template: "template",
	hybrid:   "hybrid",
	fiabName: "fiab-name",
	fiabIP:   "fiab-ip",
}

type createCommand struct {
	options options
}

func NewBeeCreateCommand() cli.ThelmaCommand {
	return &createCommand{}
}

func (cmd *createCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "create"
	cobraCommand.Short = "Create a new BEE from a template"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "NAME", "Required. Name for this BEE")
	cobraCommand.Flags().StringVarP(&cmd.options.template, flagNames.template, "t", "TEMPLATE", "Required. Template to use for this BEE")
	cobraCommand.Flags().BoolVar(&cmd.options.hybrid, flagNames.hybrid, false, "Set to true to create a hybrid (connected to a Fiab) environment")
	cobraCommand.Flags().StringVar(&cmd.options.fiabName, flagNames.fiabName, "FIAB", "Name of the Fiab this hybrid environment should be connected to")
	cobraCommand.Flags().StringVar(&cmd.options.fiabIP, flagNames.fiabIP, "IP", "Public IP address of the Fiab this hybrid environment should be connected to")
}

func (cmd *createCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	// validate --name
	if !ctx.CobraCommand().Flags().Changed(flagNames.name) {
		return fmt.Errorf("--%s is required", flagNames.name)
	}
	if err := validate.EnvironmentName(cmd.options.name); err != nil {
		return fmt.Errorf("--%s: %q is not a valid environment name: %v", flagNames.name, cmd.options.name, err)
	}

	// validate --template
	if !ctx.CobraCommand().Flags().Changed(flagNames.template) {
		return fmt.Errorf("--%s is required", flagNames.template)
	}

	// validate --hybrid arguments
	if cmd.options.hybrid {
		if !ctx.CobraCommand().Flags().Changed(flagNames.fiabName) || cmd.options.fiabName == "" {
			return fmt.Errorf("--%s is required for hybrid environments", flagNames.fiabName)
		}
		if !ctx.CobraCommand().Flags().Changed(flagNames.fiabIP) {
			return fmt.Errorf("--%s is required for hybrid environments", flagNames.fiabIP)
		}
		if !utils.IsIPV4Address(cmd.options.fiabIP) {
			return fmt.Errorf("--%s: %q is not a valid ipv4 address", flagNames.fiabIP, cmd.options.fiabIP)
		}
	}

	return nil
}

func (cmd *createCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}

	err = createEnv(cmd, state)
	if err != nil {
		return err
	}

	log.Info().Msgf("Created new environment %s", cmd.options.name)

	// reload state and print environment to console
	state, err = app.State()
	if err != nil {
		return err
	}
	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}
	if env == nil {
		// don't think this could ever happen, but let's provide a useful error anyway
		return fmt.Errorf("error creating environment %q: missing from state after creation", cmd.options.name)
	}
	ctx.SetOutput(views.ForTerraEnv(env))
	return nil
}

func (cmd *createCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do yet
	return nil
}

func createEnv(cmd *createCommand, state terra.State) error {
	template, err := getTemplate(state, cmd.options.template)

	if err != nil {
		return err
	}

	if cmd.options.hybrid {
		return state.Environments().CreateHybridFromTemplate(cmd.options.name, template, terra.NewFiab(cmd.options.fiabName, cmd.options.fiabIP))
	} else {
		return state.Environments().CreateFromTemplate(cmd.options.name, template)
	}
}

// return the template by the given name, or a helpful error listing valid configuration template names
func getTemplate(state terra.State, name string) (terra.Environment, error) {
	template, err := state.Environments().Get(name)
	if err != nil {
		return nil, err
	}
	if template != nil {
		return template, nil
	}

	templates, err := state.Environments().Filter(filter.Environments().HasLifecycle(terra.Template))
	if err != nil {
		return nil, err
	}

	var names []string
	for _, t := range templates {
		names = append(names, t.Name())
	}
	return nil, fmt.Errorf("--%s: no template by the name %q exists, valid templates are: %s", flagNames.template, name, strings.Join(names, ", "))
}

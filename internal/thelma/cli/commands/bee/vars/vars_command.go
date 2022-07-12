package pin

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
)

const helpMessage = `Generate variables for firecloud-develop FiaB config rendering`

type options struct {
	name string
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	name string
}{
	name: "name",
}

type varsCommand struct {
	options options
}

func NewBeeVarsCommand() cli.ThelmaCommand {
	return &varsCommand{}
}

func (cmd *varsCommand) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "vars [options]"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to generate variables for")
}

func (cmd *varsCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
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

func (cmd *varsCommand) Run(app app.ThelmaApp, _ cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}
	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}

	// We want to generate output in plain key-value format (not json or yaml). So, we print directly to stdout
	// instead of using rc.SetOutput()
	if env == nil {
		log.Info().Msgf("%s is a vanilla Fiab, not a BEE", cmd.options.name)
		fmt.Println("HYBRID_FIAB=false")
	} else {
		log.Info().Msgf("%s is a BEE", env.Name())

		fmt.Println("HYBRID_FIAB=true")

		for _, release := range env.Releases() {
			if release.IsAppRelease() {
				appRelease, wasAppRelease := release.(terra.AppRelease)
				if wasAppRelease {
					upname := strings.ToUpper(appRelease.Name())
					fmt.Printf("%s_HOST=%s\n", upname, appRelease.Host())
					fmt.Printf("%s_URL=%s\n", upname, appRelease.URL())
					fmt.Printf("%s_PORT=%d\n", upname, appRelease.Port())
					fmt.Printf("%s_PROTO=%s\n", upname, appRelease.Protocol())
				} else {
					log.Error().Msgf("%s was an App Release but failed to type-assert", release.Name())
				}
			} else {
				log.Error().Msgf("%s was not an App Release", release.Name())
			}
		}
	}

	return nil
}

func (cmd *varsCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

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

const helpMessage = `Generate variables for fiab config rendering`

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
	cobraCommand.Short = "Generate Fiab rendering variables for a BEE"
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().StringVarP(&cmd.options.name, flagNames.name, "n", "", "Required. Name of the BEE to pin")
}

func (cmd *varsCommand) PreRun(_ app.ThelmaApp, ctx cli.RunContext) error {
	flags := ctx.CobraCommand().Flags()

	// validate --name
	if !flags.Changed(flagNames.name) {
		return fmt.Errorf("--%s is required", flagNames.name)
	}

	return nil
}

func (cmd *varsCommand) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	state, err := app.State()
	if err != nil {
		return err
	}

	env, err := state.Environments().Get(cmd.options.name)
	if err != nil {
		return err
	}

	// okay here's a trick. we want to generate output in key-value format. Not JSON, not YAML.
	if env == nil {
		log.Info().Msgf("%s is a vanilla Fiab, not a BEE", cmd.options.name)
		fmt.Println("HYBRID_FIAB=false")
	} else {
		log.Info().Msgf("%s is a BEE", env.Name())

		fmt.Println("HYBRID_FIAB=true")

		for _, r := range env.Releases() {
			upname := strings.ToUpper(r.Name())
			proto := "https"
			port := 443

			if r.Name() == "opendj" {
				proto = "ldap"
				port = 389
			}

			host := fmt.Sprintf("%s.%s.preview.envs-terra.bio", r.Name(), env.Name())
			url := fmt.Sprintf("%s://%s", proto, host)

			fmt.Printf("%s_HOST=%s\n", upname, host)
			fmt.Printf("%s_URL=%s\n", upname, url)
			fmt.Printf("%s_PORT=%d\n", upname, port)
			fmt.Printf("%s_PROTO=%s\n", upname, proto)
		}
	}

	return nil
}

func (cmd *varsCommand) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	// nothing to do here
	return nil
}

func releasesByName(env terra.Environment) map[string]terra.Release {
	result := make(map[string]terra.Release)
	for _, r := range env.Releases() {
		result[r.Name()] = r
	}
	return result
}

package delete

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/filterflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const helpMessage = `Bulk-delete BEEs`

type options struct {
	autoDeleteOnly bool
	dryRun         bool
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	autoDeleteOnly string
	dryRun         string
}{
	autoDeleteOnly: "auto-delete-only",
	dryRun:         "dry-run",
}

type command struct {
	options options
	fflags  filterflags.FilterFlags
}

func NewBeesDeleteCommand() cli.ThelmaCommand {
	return &command{
		fflags: filterflags.NewFilterFlags(),
	}
}

func (cmd *command) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "delete"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().BoolVar(&cmd.options.dryRun, flagNames.dryRun, true, "Print the names of the BEEs that would be deleted without deleting them")
	cobraCommand.Flags().BoolVar(&cmd.options.autoDeleteOnly, flagNames.autoDeleteOnly, true, "Only include BEEs that are ready for automatic deletion")

	cmd.fflags.AddFlags(cobraCommand)
}

func (cmd *command) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (cmd *command) Run(app app.ThelmaApp, rc cli.RunContext) error {
	bees, err := builders.NewBees(app)
	if err != nil {
		return err
	}

	beeFilter, err := cmd.fflags.GetFilter(app)
	if err != nil {
		return err
	}
	if cmd.options.autoDeleteOnly {
		beeFilter = beeFilter.And(filter.Environments().AutoDeletable())
	}

	log.Info().Msgf("Filter: %s", beeFilter.String())

	matchingBees, err := bees.FilterBees(beeFilter)
	if err != nil {
		return err
	}

	if len(matchingBees) == 0 {
		log.Info().Msg("found no matching BEEs to delete")
		return nil
	}

	if cmd.options.dryRun {
		log.Info().Msgf("The following BEEs would be deleted (not deleting since this is a dry run):")
		view := views.SummarizeBees(matchingBees)
		rc.SetOutput(view)
		return nil
	}

	var names []string
	for _, env := range matchingBees {
		names = append(names, env.Name())
	}
	log.Info().Msgf("Preparing to delete %d BEEs: %#v", len(matchingBees), views.SummarizeBees(matchingBees))

	var deleted []terra.Environment
	for _, env := range matchingBees {
		log.Info().Msgf("Deleting %s", env.Name())
		_, err := bees.DeleteWith(env.Name(), bee.DeleteOptions{
			Unseed:     true,
			ExportLogs: true,
		})
		if err != nil {
			return fmt.Errorf("error deleting %s: %v", env.Name(), err)
		}
		deleted = append(deleted, env)
	}

	log.Info().Msgf("The following BEEs were deleted:")
	view := views.SummarizeBees(deleted)
	rc.SetOutput(view)
	return nil
}

func (cmd *command) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

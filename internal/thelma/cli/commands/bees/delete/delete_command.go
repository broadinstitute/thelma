package delete

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/bee"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/filterflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/broadinstitute/thelma/internal/thelma/clients/slack"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"strings"
	"sync"
)

const helpMessage = `Bulk-delete BEEs`

type options struct {
	autoDeleteOnly bool
	dryRun         bool
	maxParallel    int
}

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced in error messages
var flagNames = struct {
	autoDeleteOnly string
	dryRun         string
	maxParallel    string
}{
	autoDeleteOnly: "auto-delete-only",
	dryRun:         "dry-run",
	maxParallel:    "max-parallel",
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
	cobraCommand.Flags().IntVar(&cmd.options.maxParallel, flagNames.maxParallel, 1, "Number of BEEs to delete in parallel")

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

	slackClient, err := app.Clients().Slack()
	if err != nil {
		return errors.Errorf("failed to construct Slack client: %v", err)
	}

	var names []string
	for _, env := range matchingBees {
		names = append(names, env.Name())
	}
	log.Info().Msgf("Preparing to delete %d BEEs: %#v", len(matchingBees), strings.Join(names, ", "))

	var deleted []terra.Environment
	var mutex sync.Mutex

	var jobs []pool.Job
	for _, unsafe := range matchingBees {
		env := unsafe
		jobs = append(jobs, pool.Job{
			Name: env.Name(),
			Run: func(_ pool.StatusReporter) error {
				log.Info().Msgf("Deleting %s", env.Name())
				_, err := bees.DeleteWith(env.Name(), bee.DeleteOptions{
					Unseed:     true,
					ExportLogs: true,
				})

				sendSlackMessage(slackClient, env.Name(), err)

				if err != nil {
					return err
				}

				mutex.Lock()
				defer mutex.Unlock()
				deleted = append(deleted, env)
				return nil
			},
			Labels: map[string]string{
				"env": env.Name(),
			},
		})
	}

	err = pool.New(jobs, func(o *pool.Options) {
		o.NumWorkers = cmd.options.maxParallel
		o.Summarizer.Enabled = true
		o.Metrics.Enabled = true
		o.Metrics.PoolName = "bees_bulk_delete"
	}).Execute()

	if err != nil {
		return err
	}

	log.Info().Msgf("The following BEEs were deleted (output shows previous state):")
	view := views.SummarizeBees(deleted)
	rc.SetOutput(view)

	return nil
}

func (cmd *command) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func sendSlackMessage(slackClient *slack.Slack, name string, err error) {
	var markdown string
	if err == nil {
		markdown = fmt.Sprintf("Orphaned BEE %s was automatically deleted", name)
	} else {
		markdown = fmt.Sprintf("Failed to deleted orphaned BEE %s: %v", name, err)
	}

	if slackErr := slackClient.SendDevopsAlert("Bee Cleanup", markdown, err == nil); slackErr != nil {
		log.Warn().Err(slackErr).Msgf("error posting Slack alert: %v", slackErr)
	}
}

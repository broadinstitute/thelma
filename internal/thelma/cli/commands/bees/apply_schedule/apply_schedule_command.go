package apply_schedule

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/builders"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/filterflags"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/common/views"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/broadinstitute/thelma/internal/thelma/utils/schedule"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"sync"
	"time"
)

const helpMessage = `Start and stop BEEs as defined by their schedule`

type options struct {
	dryRun         bool
	start          bool
	stop           bool
	fromPast       time.Duration
	creationBuffer time.Duration
	maxParallel    int
}

var flagNames = struct {
	dryRun         string
	start          string
	stop           string
	fromPast       string
	creationBuffer string
	maxParallel    string
}{
	dryRun:         "dry-run",
	start:          "start",
	stop:           "stop",
	fromPast:       "from-past",
	creationBuffer: "creation-buffer",
	maxParallel:    "max-parallel",
}

type command struct {
	options options
	fflags  filterflags.FilterFlags
}

func NewBeesApplyScheduleCommand() cli.ThelmaCommand {
	return &command{
		fflags: filterflags.NewFilterFlags(),
	}
}

func (cmd *command) ConfigureCobra(cobraCommand *cobra.Command) {
	cobraCommand.Use = "apply-schedule"
	cobraCommand.Short = helpMessage
	cobraCommand.Long = helpMessage

	cobraCommand.Flags().BoolVar(&cmd.options.dryRun, flagNames.dryRun, true, "Print the actions that would be taken instead of also doing them")
	cobraCommand.Flags().BoolVar(&cmd.options.start, flagNames.start, true, "If start schedules should be applied")
	cobraCommand.Flags().BoolVar(&cmd.options.stop, flagNames.stop, true, "if stop schedules should be applied")
	cobraCommand.Flags().DurationVar(&cmd.options.fromPast, flagNames.fromPast, 20*time.Minute, "How far back to look for schedule transitions to apply (e.g. 5m, 1h, 30s)")
	cobraCommand.Flags().DurationVar(&cmd.options.creationBuffer, flagNames.creationBuffer, 20*time.Minute, "Ignore BEEs created in the past duration to allow uninterrupted seeding (e.g. 5m, 1h, 30s)")
	cobraCommand.Flags().IntVar(&cmd.options.maxParallel, flagNames.maxParallel, 3, "Number of BEEs to apply schedules for in parallel")

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

	matchingBees, err := bees.FilterBees(beeFilter)
	if err != nil {
		return err
	}

	if len(matchingBees) == 0 {
		log.Info().Msg("Found no matching BEEs to check scheduling for")
		return nil
	}

	now := time.Now()
	since := now.Add(-cmd.options.fromPast)

	var beesToFlip []terra.Environment
	for _, matchingBee := range matchingBees {

		// If this BEE was created within the buffer time, ignore it
		if matchingBee.CreatedAt().After(now.Add(-cmd.options.creationBuffer)) {
			log.Info().Msgf("skipping potential match %s since it was created %s ago", matchingBee.Name(), now.Sub(matchingBee.CreatedAt()).String())
			continue
		}

		wouldStop := cmd.options.stop &&
			matchingBee.OfflineScheduleBeginEnabled() &&
			schedule.CheckDailyScheduleMatch(matchingBee.OfflineScheduleBeginTime(), since, now)

		// We care about the day of the week for wouldStart, so let's keep server's local time as the default but
		// try to use the timezone that the schedule was defined in. We do this because it's possible that it might
		// be another day of the week in another timezone right now.
		location := time.Local
		if matchingBee.OfflineScheduleEndEnabled() && matchingBee.OfflineScheduleEndTime().Location() != nil {
			location = matchingBee.OfflineScheduleEndTime().Location()
		}
		wouldStart := cmd.options.start &&
			matchingBee.OfflineScheduleEndEnabled() &&
			schedule.CheckDailyScheduleMatch(matchingBee.OfflineScheduleEndTime(), since, now) &&
			(!schedule.IsWeekendDay(since.In(location)) || !schedule.IsWeekendDay(now.In(location)) || matchingBee.OfflineScheduleEndWeekends())

		if wouldStop && wouldStart {
			log.Warn().Msgf("%s would've been both stopped and started right now", matchingBee.Name())
			if slack, err := app.Clients().Slack(); err != nil {
				log.Debug().Msgf("Couldn't send slack message to the owner, errored building slack client: %v", err)
			} else if slack == nil {
				log.Debug().Msgf("Couldn't send slack message to the owner, no slack client")
			} else if cmd.options.dryRun {
				log.Debug().Msgf("Skipped sending slack message due to this being a dry run")
			} else {
				markdown := fmt.Sprintf("Hey there, your <https://broad.io/beehive/r/environment/%s|%s> BEE has a start/stop schedule with times too close together, so Thelma skipped applying the schedule just now. Thelma was using a range of %s, so try making the start and stop times at least that far apart.",
					matchingBee.Name(), matchingBee.Name(), cmd.options.fromPast.String())
				if err := slack.SendDirectMessage(matchingBee.Owner(), markdown); err != nil {
					log.Debug().Msgf("Couldn't send a slack message to %s: %v", matchingBee.Owner(), err)
					if err := slack.SendDevopsAlert("Conflicting Bee Schedule", fmt.Sprintf("BEE %s had conflicting start/stop times and an non-Slackable owner; try making the times at least %s apart", matchingBee.Name(), cmd.options.fromPast), false); err != nil {
						log.Debug().Msgf("Couldn't send a devops slack alert: %v", err)
					} else {
						log.Debug().Msgf("Successfully send a devops slack alert")
					}
				} else {
					log.Debug().Msgf("Successfully sent a slack message to %s", matchingBee.Owner())
				}
			}
		} else if (!matchingBee.Offline() && wouldStop) || (matchingBee.Offline() && wouldStart) {
			beesToFlip = append(beesToFlip, matchingBee)
		}
	}

	if len(beesToFlip) == 0 {
		log.Info().Msg("Found no matching BEEs with scheduling that needed applying")
		return nil
	}

	state, err := app.State()
	if err != nil {
		return err
	}

	var successfullyFlippedEnvNames []string
	var mutex sync.Mutex

	var jobs []pool.Job
	for _, unsafe := range beesToFlip {
		env := unsafe
		if cmd.options.dryRun {
			log.Info().Msgf("Would've flipped %s target state to offline=%t but dry run was enabled", env.Name(), !env.Offline())
		} else if err := state.Environments().SetOffline(env.Name(), !env.Offline()); err != nil {
			log.Warn().Msgf("Failed to flip %s target state to offline=%t: %v", env.Name(), !env.Offline(), err)
			continue
		} else {
			log.Info().Msgf("Flipped %s target state to offline=%t", env.Name(), !env.Offline())
		}
		jobs = append(jobs, pool.Job{
			Name: env.Name(),
			Run: func(_ pool.StatusReporter) error {
				if cmd.options.dryRun {
					log.Info().Msgf("Would've synced %s but dry run was enabled", env.Name())
				} else {
					log.Info().Msgf("Syncing %s", env.Name())
					_, err := bees.SyncArgoAppsIn(env, func(options *argocd.SyncOptions) {
						options.SkipLegacyConfigsRestart = true
					})
					if err != nil {
						return err
					}
				}

				mutex.Lock()
				defer mutex.Unlock()
				successfullyFlippedEnvNames = append(successfullyFlippedEnvNames, env.Name())
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
		o.Metrics.Enabled = !cmd.options.dryRun // When dry-running, don't report metrics!
		o.Metrics.PoolName = "bees_apply_schedule"
	}).Execute()

	if err != nil {
		return err
	}

	state, err = app.StateLoader().Reload()
	if err != nil {
		return fmt.Errorf("flipped BEEs but couldn't reload state: %v", err)
	}

	var successfullyFlippedEnvs []terra.Environment
	for _, envName := range successfullyFlippedEnvNames {
		if env, err := state.Environments().Get(envName); err != nil {
			log.Warn().Msgf("%s was flipped but there was an error getting it from reloaded state to output: %v", envName, err)
		} else {
			successfullyFlippedEnvs = append(successfullyFlippedEnvs, env)
		}
	}

	if cmd.options.dryRun {
		log.Info().Msg("These BEEs would've had scheduling applied if dry run wasn't enabled, which would've flipped them to the opposite of the following state:")
	} else {
		log.Info().Msgf("These BEEs had scheduling applied as follows:")
	}
	view := views.SummarizeBees(successfullyFlippedEnvs)
	rc.SetOutput(view)

	return nil
}

func (cmd *command) PostRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

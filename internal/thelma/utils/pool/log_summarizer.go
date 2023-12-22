package pool

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/repeater"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

const elapsedTimeField = "t"

type LogSummarizerOptions struct {
	// Enabled if true, print a periodic summary of pool status while items are being processed. For example:
	//
	// 2/5 items processed queued=1 running=2 success=1 error=1
	// foo:    error   err="something bad happened" dur=2m30s
	// bar:    running status="downloading file" dur=30s
	// baz:    running status="uploading file" dur=1m53s
	// quux:   queued
	// blergh: success status="finished transfer"
	//
	Enabled bool
	// Interval how frequently summary messages should be printed to the log
	Interval time.Duration
	// LogLevel level at which summary messages should be logged
	LogLevel zerolog.Level
	// WorkDescription description to use in summary header (defaults to "items processed")
	WorkDescription string
	// Footer an optional footer string to include after printing the summary
	Footer string
	// MaxLineItems an optional number of maximum line items to include in the summary
	MaxLineItems int
	logger       *zerolog.Logger
}

// log is a convenience method to create a zerolog.Event based on the configured options
func (o *LogSummarizerOptions) log() *zerolog.Event {
	var logger zerolog.Logger

	if o.logger != nil {
		logger = *o.logger
	} else {
		logger = log.Logger
	}

	return logger.WithLevel(o.LogLevel)
}

func newLogSummarizer(items []workItem, options LogSummarizerOptions) repeater.Repeater {
	return repeater.New(func() {
		logPoolSummary(items, options)
	}, func(o *repeater.Options) {
		o.Enabled = options.Enabled
		o.Interval = options.Interval
	})
}

// logPoolSummary prints a summary of a pool's workItems to the log, for example:
//
// 5/23 services synced queued=2 running=17 success=4 error=1
// thurloe:          running status={message="syncing legacy configs" attempt=2} duration=20s
// sam:              running status={message="ImagePull backoff"} duration=5m20s
// rawls:            error   err="timed out waiting for healthy" duration=17m10s
// leonardo:         success status={message="healthy"} duration=23m10s
// workspacemanager: queued
// firecloudorch:    queued
// ...
func logPoolSummary(items []workItem, options LogSummarizerOptions) {
	// While this function receives all the LogSummarizerOptions for brevity,
	// it doesn't need to worry about Enabled or Interval since those are
	// passed to and handled by the repeater.

	nameWidth := 0
	counts := make(map[Phase]int)
	for _, item := range items {
		phase := item.getPhase()
		name := item.getName()
		counts[phase] = counts[phase] + 1
		if len(name) > nameWidth {
			nameWidth = len(name)
		}
	}

	// Log a message like
	// 5/23 items processed queued=2 running=17 success=4 error=1
	processed := counts[Success] + counts[Error]
	event := options.log()
	for _, phase := range []Phase{Queued, Success, Running, Error} {
		if counts[phase] > 0 {
			event.Int(phase.String(), counts[phase])
		}
	}
	event.Msgf("%d/%d %s", processed, len(items), options.WorkDescription)

	// For large batch jobs (say N=100 items), we don't want to summarize every individual item, it's too noisy.
	// So we exclude queued items from the summary, then successful, then running, until we get a summary
	// that's under N items long.
	// If there are more than 100 error'ed items, we log the first N and stop.
	excludePhases := make(map[Phase]bool)
	count := len(items)
	for _, phase := range []Phase{Queued, Success, Running} {
		if count <= options.MaxLineItems {
			break
		}
		excludePhases[phase] = true
		count -= counts[phase]
	}

	logged := 0
	for _, item := range items {
		phase := item.getPhase()

		if logged >= options.MaxLineItems {
			break
		}
		if excludePhases[phase] {
			continue
		}

		event := options.log()

		status := item.status()
		if status != nil {
			if status.Message != "" {
				event.Str("status", status.Message)
			}
			event.Str("status", status.Message)
			if len(status.Context) > 0 {
				for k, v := range status.Context {
					event.Interface(k, v)
				}
			}
		}
		if phase != Queued {
			// optimizing for humans reading the logs
			event.Str(elapsedTimeField, item.duration().Round(time.Second).String())
		}
		if item.hasErr() {
			event.Err(item.getErr())
		}

		label := rightPad(item.getName()+":", nameWidth+1)
		event.Msgf("%s %s", label, phase.String())
		logged++
	}

	if options.Footer != "" {
		options.log().Msg(options.Footer)
	}
}

func rightPad(s string, tolen int) string {
	return fmt.Sprintf("%-*s", tolen, s)
}

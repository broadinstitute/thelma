package pool

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

const elapsedTimeField = "time"

type SummarizerOptions struct {
	// Enabled if true, print a periodic summary of pool status while items are being processed. For example:
	//
	// 2/5 items processed queued=1 running=2 success=1 error=1
	// foo:    error   err="something bad happened" duration=2m30s
	// bar:    running status="downloading file" duration=30s
	// baz:    running status="uploading file" duration=1m53s
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

func newSummarizer(items []workItem, options SummarizerOptions) *summarizer {
	return &summarizer{
		options:    options,
		items:      items,
		killSwitch: make(chan struct{}),
	}
}

// summarizer prints periodic summaries of the pool's processing to the log.
// eg.
//
// 5/23 services synced queued=2 running=17 success=4 error=1
// thurloe:          running status={message="syncing legacy configs" attempt=2} duration=20s
// sam:              running status={message="ImagePull backoff"} duration=5m20s
// rawls:            error   err="timed out waiting for healthy" duration=17m10s
// leonardo:         success status={message="healthy"} duration=23m10s
// workspacemanager: queued
// firecloudorch:    queued
// ...
type summarizer struct {
	options    SummarizerOptions
	items      []workItem
	killSwitch chan struct{}
}

func (s *summarizer) start() {
	// log initial summary before first sleep
	s.logSummary()

	go func() {
		for {
			select {
			case <-time.After(s.options.Interval):
				s.logSummary()
			case <-s.killSwitch:
				return
			}
		}
	}()
}

func (s *summarizer) stop() {
	s.killSwitch <- struct{}{}

	// log final summary
	s.logSummary()
}

func (s *summarizer) logSummary() {
	if !s.options.Enabled {
		return
	}

	nameWidth := 0
	counts := make(map[Phase]int)
	for _, item := range s.items {
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
	event := s.log()
	for _, phase := range []Phase{Queued, Success, Running, Error} {
		if counts[phase] > 0 {
			event.Int(phase.String(), counts[phase])
		}
	}
	event.Msgf("%d/%d %s", processed, len(s.items), s.options.WorkDescription)

	// For large batch jobs (say N=100 items), we don't want to summarize every individual item, it's too noisy.
	// So we exclude queued items from the summary, then successful, then running, until we get a summary
	// that's under N items long.
	// If there are more than 100 error'ed items, we log the first N and stop.
	excludePhases := make(map[Phase]bool)
	count := len(s.items)
	for _, phase := range []Phase{Queued, Success, Running} {
		if count <= s.options.MaxLineItems {
			break
		}
		excludePhases[phase] = true
		count -= counts[phase]
	}

	logged := 0
	for _, item := range s.items {
		phase := item.getPhase()

		if logged >= s.options.MaxLineItems {
			break
		}
		if excludePhases[phase] {
			continue
		}

		event := s.log()

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
			event.Str(elapsedTimeField, string(item.duration().Round(time.Second)))
		}
		if item.hasErr() {
			event.Err(item.getErr())
		}

		label := rightPad(item.getName()+":", nameWidth+1)
		event.Msgf("%s %s", label, phase.String())
		logged++
	}

	if s.options.Footer != "" {
		s.log().Msg(s.options.Footer)
	}
}

func (s *summarizer) log() *zerolog.Event {
	var logger zerolog.Logger

	if s.options.logger != nil {
		logger = *s.options.logger
	} else {
		logger = log.Logger
	}

	return logger.WithLevel(s.options.LogLevel)
}

func rightPad(s string, tolen int) string {
	return fmt.Sprintf("%-*s", tolen, s)
}

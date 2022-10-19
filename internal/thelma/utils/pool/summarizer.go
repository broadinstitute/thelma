package pool

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

const maxIndividualItemsToSummarize = 100

type SummarizerOptions struct {
	Enabled         bool
	Interval        time.Duration
	LogLevel        zerolog.Level
	WorkDescription string
	Footer          string
}

func newSummarizer(items []*workItem, options SummarizerOptions) *summarizer {
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
	items      []*workItem
	killSwitch chan struct{}
}

func (s *summarizer) start() {
	go func() {
		// log initial summary before first sleep
		s.logSummary()

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
		counts[item.phase] = counts[item.phase] + 1
		if len(item.name) > nameWidth {
			nameWidth = len(item.name)
		}
	}

	// Log a message like
	// 5/23 items processed queued=2 running=17 success=4 error=1
	processed := counts[Success] + counts[Error]
	log.WithLevel(s.options.LogLevel).
		Int(Queued.String(), counts[Queued]).
		Int(Running.String(), counts[Running]).
		Int(Success.String(), counts[Success]).
		Int(Error.String(), counts[Error]).
		Msgf("%d/%d %s", processed, len(s.items), s.options.WorkDescription)

	// For large batch jobs (say N=100 items), we don't want to summarize every individual item, it's too noisy.
	// So we exclude queued items from the summary, then successful, then running, until we get a summary
	// that's under N items long.
	// If there are more than 100 error'ed items, we log the first N and stop.
	excludePhases := make(map[Phase]bool)
	count := len(s.items)
	for _, phase := range []Phase{Queued, Success, Running} {
		if count <= maxIndividualItemsToSummarize {
			break
		}
		excludePhases[phase] = true
		count -= counts[phase]
	}

	logged := 0
	for _, item := range s.items {
		if logged >= maxIndividualItemsToSummarize {
			break
		}
		if excludePhases[item.phase] {
			continue
		}

		event := log.WithLevel(s.options.LogLevel)

		if item.status() != nil {
			event.Dict("status", item.status().Dict())
		}
		if item.phase != Queued {
			event.Dur("duration", time.Since(item.startTime))
		}

		label := rightPad(item.name+":", nameWidth+1)
		event.Msgf("%s %s", label, item.phase.String())
		logged++
	}

	if s.options.Footer != "" {
		log.Info().Msg(s.options.Footer)
	}
}

func rightPad(s string, tolen int) string {
	return fmt.Sprintf("%-*s", tolen, s)
}

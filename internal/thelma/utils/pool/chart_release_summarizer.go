package pool

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/repeater"
	"github.com/rs/zerolog/log"
	"time"
)

type ChartReleaseSummarizerOptions struct {
	Enabled  bool
	Interval time.Duration
	Do       func(map[string]string) error
}

func newChartReleaseSummarizer(items []workItem, options ChartReleaseSummarizerOptions) repeater.Repeater {
	return repeater.New(func() {
		doChartReleaseSummary(items, options.Do)
	}, func(o *repeater.Options) {
		o.Enabled = options.Do != nil && options.Enabled
		o.Interval = options.Interval
	})
}

// doChartReleaseSummary calls the "do" function with the chart release statuses.
// The chart releases will be referenced by the canonical chart release names that
// Sherlock recognizes. The statuses will be "<phase>: <status>" if there is a status
// on the workItem and just "<phase>" otherwise. Labels on the status will be ignored.
func doChartReleaseSummary(items []workItem, do func(map[string]string) error) {
	statuses := make(map[string]string, len(items))
	for _, item := range items {
		if chartReleaseName := item.getChartReleaseName(); chartReleaseName != "" {
			if status := item.status(); status != nil {
				statuses[chartReleaseName] = fmt.Sprintf("%s: %s", item.getPhase().String(), status.Message)
			} else {
				statuses[chartReleaseName] = item.getPhase().String()
			}
		}
	}
	if len(statuses) > 0 && do != nil {
		err := do(statuses)
		if err != nil {
			// We assume the "do" will do something with its errors if it actually needs to, so
			// we log all the way down at the trace level
			log.Trace().Err(err).Msg("error calling action in chart release summarizer")
		}
	}
}

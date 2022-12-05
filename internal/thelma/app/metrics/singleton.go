package metrics

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"sync"
	"time"
)

var (
	// rootAggregator the global/singleton metrics aggregator for Thelma
	// we initialize it here with a no-op aggregator
	rootAggregator = Noop()
	// initialized set to true once the singleton root aggregator has been initialized
	initialized = false
	// initializeMutex used to prevent accidental concurrent updates to the above
	initializeMutex sync.Mutex
)

// Initialize replaces the global Metrics instance with one that matches Thelma's configuration.
// It should be called once during Thelma's initialization by the Thelma builder, and nowhere else.
func Initialize(thelmaConfig config.Config, iapToken string) error {
	initializeMutex.Lock()
	defer initializeMutex.Unlock()
	if initialized {
		panic("metrics has already been initialized!")
	}

	m, err := New(thelmaConfig, iapToken)
	if err != nil {
		return err
	}

	rootAggregator = m
	initialized = true
	return nil
}

// Push pushes all metrics recorded by the root aggregator to the Thelma
// metrics gateway.
// This should ONLY be called once per Thelma run, by the Thelma root command.
func Push() error {
	return rootAggregator.push()
}

// Gauge returns a new Gauge metric. Gauges should be used for values that fluctuate over time (for example, duration of a task)
func Gauge(opts Options) GaugeMetric {
	return rootAggregator.Gauge(opts)
}

// Counter returns a new Counter metric. Counters should be used for values that accumulate over time (for example, number of times a task is executed)
func Counter(opts Options) CounterMetric {
	return rootAggregator.Counter(opts)
}

// TaskCompletion is a convenience function that records the completion of as task as both a counter
// (indicating the task completed) and a gauge (representing how long the task took to complete):
// <name>_counter
// <name>_duration_seconds
// Both metrics will include an "ok" label that will be "true" if err is nil, "false" otherwise
func TaskCompletion(opts Options, duration time.Duration, err error) {
	rootAggregator.TaskCompletion(opts, duration, err)
}

// WithLabels returns a copy of the root Metrics instance with an additional set of configured labels
func WithLabels(labels map[string]string) Metrics {
	return rootAggregator.WithLabels(labels)
}

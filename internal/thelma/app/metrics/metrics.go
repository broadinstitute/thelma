package metrics

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics/labels"
	"github.com/broadinstitute/thelma/internal/thelma/app/platform"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"
	"strconv"
	"sync"
	"time"
)

const metricNamespace = "thelma"
const jobId = "thelma"
const configKey = "metrics"

// Metrics is for aggregating metrics
type Metrics interface {
	// Gauge returns a new Gauge metric. Gauges should be used for values that fluctuate over time (for example, duration of a task)
	Gauge(opts Options) GaugeMetric
	// Counter returns a new Counter metric. Counters should be used for values that accumulate over time (for example, number of times a task is executed)
	Counter(opts Options) CounterMetric
	// TaskCompletion is a convenience function that records the completion of as task as both a counter
	// (indicating the task completed) and a gauge (representing how long the task took to complete):
	// <name>_counter
	// <name>_duration_seconds
	// Both metrics will include an "ok" label that will be "true" if err is nil, "false" otherwise
	TaskCompletion(opts Options, duration time.Duration, err error)
	// WithLabels returns a copy of this Metrics instance with an additional set of configured labels
	WithLabels(map[string]string) Metrics
	// push will upload metrics to Prometheus Gateway. Panics if metrics have already been pushed.
	push() error
}

// Options options for a metric
type Options struct {
	// Name of the metric -- will be automatically prefixed with "thelma_"
	Name string
	// Help optional help text for the metric
	Help string
	// Labels optional set of labels to apply to the metric
	Labels map[string]string
}

// CounterMetric represents a Counter metric
type CounterMetric interface {
	// Inc increments the counter
	Inc()
	// Add adds to the counter
	Add(float64)
}

// GaugeMetric represents a Gauge metric
type GaugeMetric interface {
	// Set sets the gauge to the given value
	Set(float64)
}

type metricsConfig struct {
	Enabled  bool              `default:"true"`
	PushAddr string            `default:"https://prometheus-gateway.dsp-devops.broadinstitute.org"`
	Platform platform.Platform `default:"unknown"`
}

// New returns a new Metrics instance
func New(thelmaConfig config.Config, iapToken string) (Metrics, error) {
	var cfg metricsConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return nil, err
	}

	_platform := cfg.Platform
	if _platform == platform.Unknown {
		_platform = platform.Lookup()
	}

	var pusher *push.Pusher
	if cfg.Enabled {
		log.Debug().Msgf("metrics will be pushed to %s (platform: %s)", cfg.PushAddr, _platform)
		client := newHttpClientWithBearerToken(iapToken)

		// note we use text format because it's the only kind that works with the prometheus aggregation
		// gateway
		// https://github.com/weaveworks/prom-aggregation-gateway/issues/48#issue-733152797
		pusher = push.New(cfg.PushAddr, jobId).Client(client).Format(expfmt.FmtText)
	} else {
		log.Debug().Msgf("metrics pushing is disabled")
	}

	return &metrics{
		pusher: pusher,
		// root labels added to all metrics
		labels: map[string]string{
			"platform": _platform.String(),
		},
	}, nil
}

// Noop returns a metrics instance that won't actually push metrics anywhere
func Noop() Metrics {
	return &metrics{}
}

type metrics struct {
	pusher *push.Pusher
	labels map[string]string
	mutex  sync.Mutex
	pushed bool
}

func (m *metrics) Gauge(opts Options) GaugeMetric {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        opts.Name,
		Namespace:   metricNamespace,
		Help:        opts.Help,
		ConstLabels: m.mergeAndNormalize(opts.Labels),
	})
	if m.pusher != nil {
		m.pusher.Collector(gauge)
	}
	return gauge
}

func (m *metrics) Counter(opts Options) CounterMetric {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        opts.Name,
		Namespace:   metricNamespace,
		Help:        opts.Help,
		ConstLabels: m.mergeAndNormalize(opts.Labels),
	})
	if m.pusher != nil {
		m.pusher.Collector(counter)
	}
	return counter
}

func (m *metrics) TaskCompletion(opts Options, duration time.Duration, err error) {
	okLabel := map[string]string{
		"ok": strconv.FormatBool(err == nil),
	}

	m.Counter(Options{
		Name:   opts.Name + "_count",
		Help:   opts.Help,
		Labels: labels.Merge(opts.Labels, okLabel),
	}).Inc()

	m.Gauge(Options{
		Name:   opts.Name + "_duration_seconds",
		Help:   opts.Help,
		Labels: labels.Merge(opts.Labels, okLabel),
	}).Set(duration.Seconds())
}

func (m *metrics) WithLabels(_labels map[string]string) Metrics {
	return &metrics{
		pusher: m.pusher,
		labels: m.mergeAndNormalize(_labels),
	}
}

// merge and normalize the given labels with this instance's default labels
func (m *metrics) mergeAndNormalize(extraLabels ...map[string]string) map[string]string {
	toMerge := []map[string]string{m.labels}
	toMerge = append(toMerge, extraLabels...)
	merged := labels.Merge(toMerge...)
	return labels.Normalize(merged)
}

func (m *metrics) push() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.pusher == nil {
		return nil
	}
	if m.pushed {
		panic("metrics have already been pushed")
	}
	start := time.Now()
	err := m.pusher.Push()
	if err != nil {
		return err
	}

	m.pushed = true
	duration := time.Since(start)
	log.Debug().Dur("duration", duration).Msgf("Uploaded metrics to prometheus gateway in %s", duration)
	return nil
}

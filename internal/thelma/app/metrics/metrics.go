package metrics

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"
	"time"
)

const metricNamespace = "thelma"
const jobId = "thelma"
const configKey = "metrics"

type Metrics interface {
	Gauge(opts Options) Gauge
	Counter(opts Options) Counter
	WithLabels(map[string]string) Metrics
	push() error
}

// Options options for a metric
type Options struct {
	// Name name of the metric -- will be automatically prefixed with thelma_
	Name string
	// Help optional help text for the metric
	Help string
	// Labels optional set of labels to apply to the metric
	Labels map[string]string
}

type Counter interface {
	// Inc increments the counter
	Inc()
	// Add adds to the counter
	Add(float64)
}

type Gauge interface {
	// Set sets the gauge to the given value
	Set(float64)
}

type metricsConfig struct {
	Enabled  bool     `default:"true"`
	PushAddr string   `default:"https://prometheus-gateway.dsp-devops.broadinstitute.org"`
	Platform Platform `default:"unknown"`
}

// New returns a new Metrics instance
func New(thelmaConfig config.Config, iapToken string) (Metrics, error) {
	var cfg metricsConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return nil, err
	}

	_platform := cfg.Platform
	if _platform == Unknown {
		_platform = guessPlatform()
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
			"thelma_version":  version.Version,
			"thelma_platform": _platform.String(),
		},
	}, nil
}

// Noop returns a metrics instance that won't actually push metrics anywhere
func Noop() Metrics {
	return &metrics{
		pusher: nil,
		labels: make(map[string]string),
	}
}

type metrics struct {
	pusher *push.Pusher
	labels map[string]string
}

func (m *metrics) Gauge(opts Options) Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        opts.Name,
		Namespace:   metricNamespace,
		Help:        opts.Help,
		ConstLabels: merge(m.labels, opts.Labels),
	})
	if m.pusher != nil {
		m.pusher.Collector(gauge)
	}
	return gauge
}

func (m *metrics) Counter(opts Options) Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        opts.Name,
		Namespace:   metricNamespace,
		Help:        opts.Help,
		ConstLabels: merge(m.labels, opts.Labels),
	})
	if m.pusher != nil {
		m.pusher.Collector(counter)
	}
	return counter
}

func (m *metrics) WithLabels(labels map[string]string) Metrics {
	return &metrics{
		pusher: m.pusher,
		labels: merge(m.labels, labels),
	}
}

func (m *metrics) push() error {
	if m.pusher == nil {
		return nil
	}
	start := time.Now()
	err := m.pusher.Push()
	duration := time.Since(start)
	log.Debug().Dur("duration", duration).Msgf("Uploaded metrics to prometheus gateway in %s", duration)
	return err
}

// Push pushes all metrics recorded by this metrics' pusher to the Thelma
// metrics gateway.
// This should ONLY be called once per Thelma run, by the Thelma root command.
func Push(m Metrics) error {
	return m.push()
}

// LabelsForRelease returns a standard set of labels for a chart release
func LabelsForRelease(release terra.Release) map[string]string {
	labels := make(map[string]string)
	labels["release"] = release.Name()
	labels["release_key"] = release.FullName()
	labels["release_type"] = release.Type().String()
	labels["chart"] = release.ChartName()
	labels["cluster"] = release.Cluster().Name()
	return merge(LabelsForDestination(release.Destination()), labels)
}

// LabelsForDestination returns a standard set of labels for a destination
func LabelsForDestination(dest terra.Destination) map[string]string {
	labels := make(map[string]string)
	labels["destination_type"] = dest.Type().String()
	labels["destination_name"] = dest.Name()
	if dest.IsEnvironment() {
		labels["env"] = dest.Name()
	}
	if dest.IsCluster() {
		labels["cluster"] = dest.Name()
	}
	return labels
}

// merge N maps into a single map (last takes precedence)
func merge(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		if m == nil {
			// ignore nil maps
			continue
		}
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

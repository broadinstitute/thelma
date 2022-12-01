package metrics

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"net/http"
)

const metricNamespace = "thelma"
const jobId = "thelma"
const configKey = "metrics"

type metricsConfig struct {
	Enabled  bool     `default:"false"`
	PushAddr string   `default:"https://prometheus-gateway.dsp-devops.broadinstitute.org"`
	Platform Platform `default:"unknown"`
}

func New(thelmaConfig config.Config, iapToken string) (Metrics, error) {
	var cfg metricsConfig
	if err := thelmaConfig.Unmarshal(configKey, &cfg); err != nil {
		return nil, err
	}

	transport := bearerRoundTripper{
		token: iapToken,
		inner: http.DefaultTransport,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   0,
	}

	var pusher *push.Pusher
	if cfg.Enabled {
		pusher = push.New(cfg.PushAddr, jobId).Format(expfmt.FmtText).Client(client)
	}

	return &metrics{
		pusher: pusher,
		// root labels added to all metrics
		labels: map[string]string{
			"thelma_version":  version.Version,
			"thelma_platform": cfg.Platform.String(),
		},
	}, nil
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
	return m.pusher.Push()
}

type Options struct {
	Name   string
	Help   string
	Labels map[string]string
}

type Counter interface {
	// Inc increments the counter
	Inc()
}

type Gauge interface {
	// Set sets the gauge to the given value
	Set(float64)
}

type Metrics interface {
	Gauge(opts Options) Gauge
	Counter(opts Options) Counter
	WithLabels(map[string]string) Metrics
	push() error
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

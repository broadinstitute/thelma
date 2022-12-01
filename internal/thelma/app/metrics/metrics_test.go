package metrics

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const token = `eyJhbGciOiJSUzI1NiIsImtpZCI6IjE4MzkyM2M4Y2ZlYzEwZjkyY2IwMTNkMDZlMWU3Y2RkNzg3NGFlYTUiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiIxMDM4NDg0ODk0NTg1LWs4cXZmN2w4NzY3MzNsYWV2MGxtOGtlbmZhMmxqNmJuLmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29tIiwiYXVkIjoiMTAzODQ4NDg5NDU4NS1rOHF2ZjdsODc2NzMzbGFldjBsbThrZW5mYTJsajZibi5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsInN1YiI6IjExMTU3MTg0ODA1ODQxNDQ2MTIzNCIsImhkIjoiYnJvYWRpbnN0aXR1dGUub3JnIiwiZW1haWwiOiJjaGVsc2VhQGJyb2FkaW5zdGl0dXRlLm9yZyIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJhdF9oYXNoIjoiZjRwa09BdDBqT25scmpxNWFiRnJRQSIsIm5hbWUiOiJDaGVsc2VhIEhvb3ZlciIsInBpY3R1cmUiOiJodHRwczovL2xoMy5nb29nbGV1c2VyY29udGVudC5jb20vYS9BTG01d3UxazEtWExCUnhaelJUbmhIcnljUFZaODdvOFV5M2JrX1AxY1E5ND1zOTYtYyIsImdpdmVuX25hbWUiOiJDaGVsc2VhIiwiZmFtaWx5X25hbWUiOiJIb292ZXIiLCJsb2NhbGUiOiJlbiIsImlhdCI6MTY2OTkxMTQxNiwiZXhwIjoxNjY5OTE1MDE2fQ.VcG2V7WXkRg4KSXy0_DQi4QwQj32JcMLa_sRN_OChrW0m8x92Dvbhd23vH2qZuSyZ-PnwxykS_ZsGEItqhD0l1IMOcziohLESxKX-O9TAcqoZ8pXzAUNTrD7zO35v8PMzbP6WhGakqwkcTdFfHhJ6FDCl6LdWKGJRBS1dx_ZkKLuhPDosCRH7drdj3X1Kteig4tJ32wHJwWRV1WRYBCvTkKPkfNcQWi5xPjWW9v_m1zg-k1lbsaWAVHzviX-tG9VMRqfwEigUS8bCsf0qmLEgohZvtoxsgXqeS0mgA8AHsrFI8aUjT1ak652kZKbBQfINkUmyUkCCmoUkTzJk1LUMg`

func Test_API(t *testing.T) {
	thelmaConfig, err := config.NewTestConfig(t, map[string]interface{}{
		"metrics.enabled": true,
	})
	require.NoError(t, err)

	m, err := New(thelmaConfig, token)
	require.NoError(t, err)

	gauge := m.Gauge(Options{
		Name:   "test_gauge",
		Help:   "A test gauge",
		Labels: nil,
	})

	gauge.Set(12)

	counter := m.Counter(Options{
		Name:   "test_counter",
		Help:   "A test counter",
		Labels: map[string]string{"foo": "bar"},
	})

	for i := 0; i < 10; i++ {
		counter.Inc()
	}

	require.NoError(t, m.push())
}

func Test_MetricsExample(t *testing.T) {
	completionTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_backup_last_completion_timestamp_seconds",
		Help: "The timestamp of the last successful completion of a DB backup.",
		ConstLabels: map[string]string{
			"abc": "def",
		},
	})
	completionTime.SetToCurrentTime()

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "my_test_counter",
		Help: "A test counter",
		ConstLabels: map[string]string{
			"foo": "bar",
		},
	})

	hist := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "my_example_histogram",
		Help: "This is an example histogram",
		ConstLabels: map[string]string{
			"aaa": "bbb",
		},
		Buckets: []float64{},
	})
	hist.Observe(0.045)

	if err := push.New("http://localhost:80", "ping_test").
		Collector(completionTime).
		Collector(counter).
		Collector(hist).
		Format(expfmt.FmtText).
		Push(); err != nil {
		fmt.Println("Could not push completion time to Pushgateway:", err)
	}
}

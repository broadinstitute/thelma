package metrics

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_Metrics(t *testing.T) {
	fakeToken := "my-fake-token"

	testCases := []struct {
		name     string
		record   func(m Metrics)
		expected string
	}{
		{
			name: "basic counter",
			record: func(m Metrics) {
				// make sure that basic counters and gauges both show up with default dimensions
				m.Counter(Options{
					Name: "basic_counter",
					Help: "A test counter",
				}).Add(42)
			},
			expected: `
# HELP thelma_basic_counter A test counter
# TYPE thelma_basic_counter counter
thelma_basic_counter{platform="jenkins"} 42`,
		},
		{
			name: "basic gauge with custom label",
			record: func(m Metrics) {
				m.Gauge(Options{
					Name:   "basic_gauge",
					Labels: map[string]string{"foo": "bar"},
				}).Set(100)
			},
			expected: `
# HELP thelma_basic_gauge 
# TYPE thelma_basic_gauge gauge
thelma_basic_gauge{foo="bar",platform="jenkins"} 100`,
		},
		{
			name: "inherit labels",
			record: func(m Metrics) {
				withLabels := m.
					WithLabels(map[string]string{"a": "1"}).
					WithLabels(map[string]string{"b": "2"}).
					WithLabels(map[string]string{"c": "3"})

				withLabels.Gauge(Options{
					Name: "inherit_gauge",
				}).Set(-100)

				withLabels.Counter(Options{
					Name:   "inherit_counter",
					Labels: map[string]string{"baz": "quux", "b": "43"},
				}).Add(24)
			},
			expected: `
# HELP thelma_inherit_counter 
# TYPE thelma_inherit_counter counter
thelma_inherit_counter{a="1",b="43",baz="quux",c="3",platform="jenkins"} 24
# HELP thelma_inherit_gauge 
# TYPE thelma_inherit_gauge gauge
thelma_inherit_gauge{a="1",b="2",c="3",platform="jenkins"} -100`,
		},
		{
			name: "mutate metrics",
			record: func(m Metrics) {
				counter := m.Counter(Options{
					Name: "incremented_counter",
				})
				counter.Inc()
				counter.Inc()
				counter.Inc()

				gauge := m.Gauge(Options{
					Name: "updated_gauge",
				})
				gauge.Set(0.5)
				gauge.Set(0.4)
				gauge.Set(0.3)
			},
			expected: `
# HELP thelma_incremented_counter 
# TYPE thelma_incremented_counter counter
thelma_incremented_counter{platform="jenkins"} 3
# HELP thelma_updated_gauge 
# TYPE thelma_updated_gauge gauge
thelma_updated_gauge{platform="jenkins"} 0.3`,
		},
		{
			name: "reserved label should be normalized",
			record: func(m Metrics) {
				m.Counter(Options{
					Name:   "counter",
					Labels: map[string]string{"a": "1", "job": "fixme", "b": "2"},
				}).Add(123)
			},
			expected: `
# HELP thelma_counter 
# TYPE thelma_counter counter
thelma_counter{_job="fixme",a="1",b="2",platform="jenkins"} 123
`,
		},
		{
			name: "task completion helper",
			record: func(m Metrics) {

				// test task completion success
				m.TaskCompletion(Options{
					Name:   "fake_task",
					Help:   "A fake task",
					Labels: map[string]string{"id": "23"},
				}, 1200*time.Millisecond, nil)

				// test task completion failure
				m.TaskCompletion(Options{
					Name:   "fake_task",
					Help:   "A fake task",
					Labels: map[string]string{"id": "23"},
				}, 1800*time.Millisecond, fmt.Errorf("this one failed"))

			},
			expected: `# HELP thelma_fake_task_count A fake task
# TYPE thelma_fake_task_count counter
thelma_fake_task_count{id="23",ok="false",platform="jenkins"} 1
thelma_fake_task_count{id="23",ok="true",platform="jenkins"} 1
# HELP thelma_fake_task_duration_seconds A fake task
# TYPE thelma_fake_task_duration_seconds gauge
thelma_fake_task_duration_seconds{id="23",ok="false",platform="jenkins"} 1.8
thelma_fake_task_duration_seconds{id="23",ok="true",platform="jenkins"} 1.2
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var requestCount int
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// verify request has correct bearer token
				assert.Equal(t, "Bearer "+fakeToken, r.Header.Get("Authorization"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(string(body)))
				w.WriteHeader(200)

				// count requestCount so we can verify request was actually received by test server
				requestCount++
			}))
			t.Cleanup(func() {
				svr.Close()
			})

			thelmaConfig, err := config.NewTestConfig(t, map[string]interface{}{
				"metrics.enabled":  true,
				"metrics.pushaddr": svr.URL,
				"metrics.platform": "jenkins",
			})
			require.NoError(t, err)

			_metrics, err := New(thelmaConfig, fakeToken)
			require.NoError(t, err)

			tc.record(_metrics)

			require.NoError(t, err, _metrics.push())
			assert.Equal(t, 1, requestCount)
		})
	}
}

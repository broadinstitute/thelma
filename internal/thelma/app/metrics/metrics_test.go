package metrics

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_Metrics(t *testing.T) {
	fakeToken := "my-fake-token"
	expectedPushRequest := `
# HELP thelma_counter1 A test counter
# TYPE thelma_counter1 counter
thelma_counter1{thelma_platform="jenkins",thelma_version="unset"} 42
# HELP thelma_counter2 
# TYPE thelma_counter2 counter
thelma_counter2{a="1",b="2",baz="quux",thelma_platform="jenkins",thelma_version="unset"} 3
# HELP thelma_gauge1 
# TYPE thelma_gauge1 gauge
thelma_gauge1{foo="bar",thelma_platform="jenkins",thelma_version="unset"} 100
# HELP thelma_gauge2 
# TYPE thelma_gauge2 gauge
thelma_gauge2{a="1",b="2",city="nyc",thelma_platform="jenkins",thelma_version="unset",x="y"} 3.14
`

	var requestCount int
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify request has correct bearer token
		assert.Equal(t, "Bearer "+fakeToken, r.Header.Get("Authorization"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.Equal(t, strings.TrimSpace(expectedPushRequest), strings.TrimSpace(string(body)))
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

	// make sure that basic counters and gauges both show up with default dimensions
	_metrics.Counter(Options{
		Name: "counter1",
		Help: "A test counter",
	}).Add(42)

	_metrics.Gauge(Options{
		Name:   "gauge1",
		Labels: map[string]string{"foo": "bar"},
	}).Set(100)

	// now make sure that views with custom labels appear
	_metrics2 := _metrics.WithLabels(map[string]string{"a": "1", "b": "2"})
	counter2 := _metrics2.Counter(Options{
		Name:   "counter2",
		Labels: map[string]string{"baz": "quux"},
	})

	// make sure incrementing counters works
	counter2.Inc()
	counter2.Inc()
	counter2.Inc()

	// now make sure that labels are inherited
	_metrics3 := _metrics2.WithLabels(map[string]string{"x": "y"})
	_metrics3.Gauge(Options{
		Name:   "gauge2",
		Labels: map[string]string{"city": "nyc"},
	}).Set(3.14)

	require.NoError(t, err, _metrics.push())
	assert.Equal(t, 1, requestCount)
}

func Test_merge(t *testing.T) {
	assert.Equal(t, map[string]string{}, merge())
	assert.Equal(t, map[string]string{"a": "b"}, merge(map[string]string{"a": "b"}))
	assert.Equal(t, map[string]string{"a": "c"}, merge(map[string]string{"a": "b"}, map[string]string{"a": "c"}))
	assert.Equal(t, map[string]string{"a": "b"}, merge(map[string]string{"a": "c"}, map[string]string{"a": "b"}))
	assert.Equal(t, map[string]string{"a": "d"}, merge(map[string]string{"a": "b"}, map[string]string{"a": "c"}, map[string]string{"a": "d"}))

	assert.Equal(t, map[string]string{"a": "d"}, merge(nil, map[string]string{"a": "c"}, map[string]string{"a": "d"}))
	assert.Equal(t, map[string]string{"a": "c"}, merge(nil, map[string]string{"a": "c"}))
}

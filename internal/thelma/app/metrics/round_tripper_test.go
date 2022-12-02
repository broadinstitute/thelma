package metrics

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_RoundTripper(t *testing.T) {
	fakeToken := "my-fake-token"

	var requests int
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify request has correct bearer token
		assert.Equal(t, "Bearer "+fakeToken, r.Header.Get("Authorization"))
		w.WriteHeader(200)

		// count requests so we can verify request was actually received by test server
		requests++
	}))
	t.Cleanup(func() {
		svr.Close()
	})

	// create new client with the fake token
	client := newHttpClientWithBearerToken(fakeToken)

	// use the client to make a request -- server should verify request has the correct header
	_, err := client.Get(svr.URL)
	require.NoError(t, err)

	// make sure the server actually received a request
	assert.Equal(t, 1, requests)
}

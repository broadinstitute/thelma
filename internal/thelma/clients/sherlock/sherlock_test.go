package sherlock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/stretchr/testify/require"
)

type mockOkResponse struct {
	Ok bool
}

func Test_NewClient(t *testing.T) {
	mockSherlockServer := httptest.NewServer(newMockSherlockHandler())
	defer mockSherlockServer.Close()

	thelmaConfig, err := config.Load(config.WithTestDefaults(t), config.WithOverride("sherlock.addr", mockSherlockServer.URL))
	require.NoError(t, err)

	client, err := New(thelmaConfig, "fake")
	require.NoError(t, err)

	err = client.getStatus()
	require.NoError(t, err)
}

func newMockSherlockHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOkResponse{Ok: true})
	})
}

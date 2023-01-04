package terraapi

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googleoauth "google.golang.org/api/oauth2/v2"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_OrchClientRetriesFailedRequests(t *testing.T) {
	var requestCount int

	fakeOrchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/register/profile" {
			t.Errorf("expected to request '/register/profile', got: %s", r.URL.Path)
		}

		var status int
		var body string

		if requestCount > 3 {
			status = http.StatusInternalServerError
			body = `this error does not match our retryable list, so should not cause a retry`
		} else {
			status = http.StatusInternalServerError
			body = `error connecting to thurloe: java.net.UnknownHostException`
		}

		w.WriteHeader(status)
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("error writing fake http response: %v", err)
		}
	}))
	defer fakeOrchServer.Close()

	orchRelease := mocks.NewAppRelease(t)
	orchRelease.On("URL").Return(fakeOrchServer.URL)

	client := &firecloudOrchClient{
		terraClient: &terraClient{
			tokenSource: testutils.NewFakeTokenSource("fake-token"),
			userInfo:    googleoauth.Userinfo{},
			httpClient:  *fakeOrchServer.Client(),
		},
		appRelease: orchRelease,
		retryDelay: 5 * time.Millisecond,
	}

	_, _, err := client.RegisterProfile(
		"Jane", "Doe",
		"Owner", "jdoe@broadinstitute.org",
		"None", "None",
		"None", "None", "None",
		"None", "None",
	)
	require.Error(t, err)
	assert.ErrorContains(t, err, "does not match our retryable list")
	assert.Equal(t, 4, requestCount)
}

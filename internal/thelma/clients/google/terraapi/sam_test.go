package terraapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googleoauth "google.golang.org/api/oauth2/v2"
)

func Test_SamClientRetriesFailedRequests(t *testing.T) {
	testCases := []struct {
		name              string
		retryableErrCount int
		finalErr          string
		expectErr         bool
	}{
		{
			name:              "no error",
			retryableErrCount: 0,
			expectErr:         false,
		},
		{
			name:              "1 non-retryable err",
			retryableErrCount: 0,
			finalErr:          "something weird happened",
			expectErr:         true,
		},
		{
			name:              "3 retryable, 1 non-retryable err",
			retryableErrCount: 3,
			finalErr:          "something weird happened",
			expectErr:         true,
		},
		{
			name:              "5 retryable, then success",
			retryableErrCount: 5,
			expectErr:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requestCount int

			fakeSamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				if r.URL.Path != "/register/user/v1/termsofservice" {
					t.Errorf("expected to request '/register/profile', got: %s", r.URL.Path)
				}

				var status int
				var body string

				if requestCount > tc.retryableErrCount {
					if tc.finalErr != "" {
						status = http.StatusConflict
						body = "409 Conflict: this error should not be retried: " + tc.finalErr
					} else {
						status = http.StatusCreated
						body = "Created"
					}
				} else {
					status = http.StatusInternalServerError
					body = `This error is retryable`
				}

				w.WriteHeader(status)
				if _, err := w.Write([]byte(body)); err != nil {
					t.Errorf("error writing fake http response: %v", err)
				}
			}))
			defer fakeSamServer.Close()

			samRelease := mocks.NewAppRelease(t)
			samRelease.On("URL").Return(fakeSamServer.URL)

			client := &samClient{
				terraClient: &terraClient{
					tokenSource: testutils.NewFakeTokenSource("fake-token"),
					userInfo:    googleoauth.Userinfo{},
					httpClient:  *fakeSamServer.Client(),
					retryDelay:  5 * time.Millisecond,
				},
				appRelease: samRelease,
			}

			_, _, err := client.AcceptToS()
			assert.Equal(t, tc.retryableErrCount+1, requestCount)

			if tc.expectErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.finalErr)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

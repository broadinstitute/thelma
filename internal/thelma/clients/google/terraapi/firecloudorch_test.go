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

func Test_OrchClientRetriesFailedRequests(t *testing.T) {
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
		// Treat all errors as retryable for now, there are just too many different errors that can occur
		// {
		// 	name:              "1 non-retryable err",
		// 	retryableErrCount: 0,
		// 	finalErr:          "should not cause a retry",
		// 	expectErr:         true,
		// },
		// {
		// 	name:              "3 retryable, 1 non-retryable err",
		// 	retryableErrCount: 3,
		// 	finalErr:          "should not cause a retry",
		// 	expectErr:         true,
		// },
		{
			name:              "5 retryable, then success",
			retryableErrCount: 5,
			expectErr:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requestCount int

			fakeOrchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				if r.URL.Path != "/register/profile" {
					t.Errorf("expected to request '/register/profile', got: %s", r.URL.Path)
				}

				var status int
				var body string

				if requestCount > tc.retryableErrCount {
					if tc.finalErr != "" {
						status = http.StatusInternalServerError
						body = tc.finalErr
					} else {
						status = http.StatusCreated
						body = "Created"
					}
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

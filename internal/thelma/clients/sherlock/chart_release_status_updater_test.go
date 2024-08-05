package sherlock

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/sherlock/sherlock-go-client/client/models"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChartReleaseUpdater(t *testing.T) {
	t.Run("gha oidc unhappy", func(t *testing.T) {
		mockSherlockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, fmt.Sprintf("Bearer %s", testIapToken), r.Header.Get("Authorization"))
			t.Fail()
		}))
		defer mockSherlockServer.Close()
		client, err := NewClient(func(options *Options) {
			options.Addr = mockSherlockServer.URL
			// No GHA OIDC token provider
		})
		require.NoError(t, err)
		err = client.UpdateChartReleaseStatuses(map[string]string{"foo": "bar"})
		require.NoError(t, err)
	})
	t.Run("sherlock returns error", func(t *testing.T) {
		mockSherlockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body models.SherlockCiRunV3Upsert
			require.Equal(t, fmt.Sprintf("Bearer %s", testIapToken), r.Header.Get("Authorization"))
			require.Equal(t, testGhaToken, r.Header.Get(sherlockGhaOidcHeader))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, models.SherlockCiRunV3Upsert{
				ChartReleaseStatuses: map[string]string{"foo": "bar"},
			}, body)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			require.NoError(t, json.NewEncoder(w).Encode(models.ErrorsErrorResponse{
				Message: "foo not found",
				ToBlame: "client",
				Type:    "HTTP Not Found",
			}))
		}))
		defer mockSherlockServer.Close()
		client, err := NewClient(func(options *Options) {
			options.Addr = mockSherlockServer.URL
			options.IapTokenProvider = &credentials.MockTokenProvider{ReturnString: testIapToken}
			options.GhaOidcTokenProvider = &credentials.MockTokenProvider{ReturnString: testGhaToken}
		})
		require.NoError(t, err)
		err = client.UpdateChartReleaseStatuses(map[string]string{"foo": "bar"})
		require.ErrorContains(t, err, "foo not found")
	})
	t.Run("sherlock returns nothing", func(t *testing.T) {
		mockSherlockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body models.SherlockCiRunV3Upsert
			require.Equal(t, fmt.Sprintf("Bearer %s", testIapToken), r.Header.Get("Authorization"))
			require.Equal(t, testGhaToken, r.Header.Get(sherlockGhaOidcHeader))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, models.SherlockCiRunV3Upsert{
				ChartReleaseStatuses: map[string]string{"foo": "bar"},
			}, body)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			require.NoError(t, json.NewEncoder(w).Encode(models.SherlockCiRunV3{
				ID: 123,
				RelatedResources: []*models.SherlockCiIdentifierV3{
					{
						ResourceType:   "chart-release",
						ResourceID:     111,
						ResourceStatus: "bar",
					},
					{
						ResourceType:   "changeset",
						ResourceID:     222,
						ResourceStatus: "bar",
					},
					{
						ResourceType: "environment",
						ResourceID:   333,
					},
				},
			}))
		}))
		defer mockSherlockServer.Close()
		client, err := NewClient(func(options *Options) {
			options.Addr = mockSherlockServer.URL
			options.IapTokenProvider = &credentials.MockTokenProvider{ReturnString: testIapToken}
			options.GhaOidcTokenProvider = &credentials.MockTokenProvider{ReturnString: testGhaToken}
		})
		require.NoError(t, err)
		err = client.UpdateChartReleaseStatuses(map[string]string{"foo": "bar"})
		require.NoError(t, err)
	})
}

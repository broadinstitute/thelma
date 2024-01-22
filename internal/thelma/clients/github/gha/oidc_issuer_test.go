package gha

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_getOidcRequestValues(t *testing.T) {
	t.Run("both missing", func(t *testing.T) {
		_, _, err := getOidcRequestValues()
		require.Error(t, err)
	})
	t.Run("url missing", func(t *testing.T) {
		utils.OverrideEnvVarForTest(t, ghaOidcRequestTokenEnvVar, "test token", func() {
			_, _, err := getOidcRequestValues()
			require.ErrorContains(t, err, ghaOidcRequestUrlEnvVar)
		})
	})
	t.Run("token missing", func(t *testing.T) {
		utils.OverrideEnvVarForTest(t, ghaOidcRequestUrlEnvVar, "test url", func() {
			_, _, err := getOidcRequestValues()
			require.ErrorContains(t, err, ghaOidcRequestTokenEnvVar)
		})
	})
	t.Run("both present", func(t *testing.T) {
		utils.OverrideEnvVarForTest(t, ghaOidcRequestUrlEnvVar, "test url", func() {
			utils.OverrideEnvVarForTest(t, ghaOidcRequestTokenEnvVar, "test token", func() {
				url, token, err := getOidcRequestValues()
				require.Equal(t, "test url", url)
				require.Equal(t, "test token", token)
				require.NoError(t, err)
			})
		})
	})
	t.Run("url empty", func(t *testing.T) {
		utils.OverrideEnvVarForTest(t, ghaOidcRequestUrlEnvVar, "", func() {
			utils.OverrideEnvVarForTest(t, ghaOidcRequestTokenEnvVar, "test token", func() {
				_, _, err := getOidcRequestValues()
				require.ErrorContains(t, err, ghaOidcRequestUrlEnvVar)
			})
		})
	})
	t.Run("token empty", func(t *testing.T) {
		utils.OverrideEnvVarForTest(t, ghaOidcRequestUrlEnvVar, "test url", func() {
			utils.OverrideEnvVarForTest(t, ghaOidcRequestTokenEnvVar, "", func() {
				_, _, err := getOidcRequestValues()
				require.ErrorContains(t, err, ghaOidcRequestTokenEnvVar)
			})
		})
	})
}

func Test_getOidcToken(t *testing.T) {
	var mockStatusCode int
	mockRequestToken := "mock request token"
	mockOidcResponseToken := "something we treat as []byte"
	mockIssuer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, fmt.Sprintf("Bearer %s", mockRequestToken), r.Header.Get("Authorization"))
		w.WriteHeader(mockStatusCode)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"value": mockOidcResponseToken,
		}))
	}))
	defer mockIssuer.Close()

	utils.OverrideEnvVarForTest(t, ghaOidcRequestUrlEnvVar, mockIssuer.URL, func() {
		utils.OverrideEnvVarForTest(t, ghaOidcRequestTokenEnvVar, mockRequestToken, func() {
			for _, statusCode := range []int{http.StatusTeapot, http.StatusOK} {
				mockStatusCode = statusCode
				tokenBytes, err := getOidcToken()
				require.NoError(t, err)
				require.Equal(t, mockOidcResponseToken, string(tokenBytes))
			}
		})
	})
}

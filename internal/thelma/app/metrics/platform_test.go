package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Platform_Serializers(t *testing.T) {

	testCases := []struct {
		input     string
		expected  Platform
		expectErr string
	}{
		{
			input:    "unknown",
			expected: Unknown,
		},
		{
			input:    "local",
			expected: Local,
		},
		{
			input:    "argocd",
			expected: ArgoCD,
		},
		{
			input:    "gha",
			expected: GithubActions,
		},
		{
			input:    "jenkins",
			expected: Jenkins,
		},
		{
			input:     "raise err",
			expectErr: "invalid platform",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			var p Platform

			// convert input string into quoted string (so, valid json)
			asJson := fmt.Sprintf("%q", tc.input)
			err := json.Unmarshal([]byte(asJson), &p)
			if tc.expectErr != "" {
				require.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			// make sure string was deserialized into correct platform type
			assert.Equal(t, tc.expected, p)
			// make sure stringer returns the input string
			assert.Equal(t, tc.input, p.String())
		})
	}
}

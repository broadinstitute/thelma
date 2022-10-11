package pinflags

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NormalizeImageTag(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "",
		},
		{
			input:    "foo",
			expected: "foo",
		},
		{
			input:    ".",
			expected: "",
		},
		{
			input:    "-",
			expected: "",
		},
		{
			input:    "a.",
			expected: "a.",
		},
		{
			input:    "a-",
			expected: "a-",
		},
		{
			input:    "/with/slashes/",
			expected: "with-slashes-",
		},
		{
			input:    "has?ill*egal)chars",
			expected: "has-ill-egal-chars",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, normalizeImageTag(tc.input))
	}
}

func Test_NormalizeImageTags(t *testing.T) {
	input := map[string]terra.VersionOverride{
		"sam": {
			AppVersion:          "/with/sl(ash)es/",
			ChartVersion:        "1.2.3",
			TerraHelmfileRef:    "master",
			FirecloudDevelopRef: "dev",
		},
	}

	expected := map[string]terra.VersionOverride{
		"sam": {
			AppVersion:          "with-sl-ash-es-",
			ChartVersion:        "1.2.3",
			TerraHelmfileRef:    "master",
			FirecloudDevelopRef: "dev",
		},
	}

	assert.Equal(t, expected, normalizeImageTags(input))
}

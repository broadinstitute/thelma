package format

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var data = struct {
	Foo string `json:"foo"`
}{
	Foo: "bar",
}

func TestFormat(t *testing.T) {
	testCases := []struct {
		name     string
		format   Format
		expected string
	}{
		{
			name:     "yaml",
			format:   Yaml,
			expected: "foo: bar\n",
		},
		{
			name:     "json",
			format:   Json,
			expected: "{\n  \"foo\": \"bar\"\n}\n",
		},
		{
			name:     "none",
			format:   None,
			expected: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			require.NoError(t, tc.format.Format(data, &b))
			assert.Equal(t, tc.expected, b.String())
		})
	}
}

func TestPretty(t *testing.T) {
	var b bytes.Buffer
	require.NoError(t, formatPrettyYamlWithOptions(data, &b, "fruity", "terminal16m"))
	assert.Equal(t, "\x1b[1m\x1b[38;2;251;102;10mfoo\x1b[0m\x1b[38;2;255;255;255m:\x1b[0m\x1b[38;2;136;136;136m \x1b[0m\x1b[38;2;255;255;255mbar\x1b[0m\x1b[38;2;136;136;136m\n\x1b[0m", b.String())
}

func TestIsSupported(t *testing.T) {
	assert.True(t, IsSupported("yaml"))
	assert.False(t, IsSupported("foo"))
}

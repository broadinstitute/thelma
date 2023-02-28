package bootstrap

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_computeTerminalWidth(t *testing.T) {
	testCases := []struct {
		input    int
		expected int
	}{
		{
			input:    0,
			expected: 8,
		},
		{
			input:    10,
			expected: 0,
		},
		{
			input:    50,
			expected: 2,
		},
		{
			input:    80,
			expected: 17,
		},
		{
			input:    81,
			expected: 18,
		},
		{
			input:    82,
			expected: 18,
		},
		{
			input:    120,
			expected: 37,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d", tc.input), func(t *testing.T) {
			assert.Equal(t, tc.expected, computeLeftPaddingToCenterLogo(tc.input))
		})
	}
}

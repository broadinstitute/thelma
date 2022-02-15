package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CommandKey(t *testing.T) {
	type expected struct {
		depth       int
		isRoot      bool
		ancestors   []string
		shortName   string
		longName    string
		description string
	}

	testCases := []struct {
		name     string
		input    commandKey
		testFn   func(t *testing.T, input commandKey)
		expected expected
	}{
		{
			name:  "root",
			input: rootCommandKey(),
			expected: expected{
				depth:       0,
				isRoot:      true,
				ancestors:   []string{},
				shortName:   "",
				longName:    "",
				description: "root",
			},
		},
		{
			name:  "one level",
			input: newCommandKey("foo"),
			expected: expected{
				depth:       1,
				isRoot:      false,
				ancestors:   []string{},
				shortName:   "foo",
				longName:    "foo",
				description: "foo",
			},
		},
		{
			name:  "two levels",
			input: newCommandKey("foo bar"),
			expected: expected{
				depth:       2,
				isRoot:      false,
				ancestors:   []string{"foo"},
				shortName:   "bar",
				longName:    "foo bar",
				description: "foo bar",
			},
		},

		{
			name:  "three levels",
			input: newCommandKey("foo bar baz"),
			expected: expected{
				depth:       3,
				isRoot:      false,
				ancestors:   []string{"foo", "bar"},
				shortName:   "baz",
				longName:    "foo bar baz",
				description: "foo bar baz",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected.depth, tc.input.depth())
			assert.Equal(t, tc.expected.isRoot, tc.input.isRoot())
			assert.Equal(t, tc.expected.ancestors, tc.input.ancestors())
			assert.Equal(t, tc.expected.shortName, tc.input.shortName())
			assert.Equal(t, tc.expected.longName, tc.input.longName())
			assert.Equal(t, tc.expected.description, tc.input.description())
		})
	}
}

package cli

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	testCases := []struct {
		name                 string
		thelmaArgs           string
		expectErrorMatching  string
		expectOutputMatching string
	}{
		{
			name:                "extra arguments should return an error",
			thelmaArgs:          "version foo bar",
			expectErrorMatching: `expected 0 arguments, got 2: \[foo bar\]`,
		},
		{
			name:                 "should print version",
			thelmaArgs:           "version",
			expectOutputMatching: "version: unset\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			thelmaCLI := newThelmaCLI()
			var stdout bytes.Buffer
			thelmaCLI.setArgs(strings.Fields(tc.thelmaArgs))
			thelmaCLI.setStdout(&stdout)

			err := thelmaCLI.rootCommand.Execute()
			if tc.expectErrorMatching != "" {
				assert.Error(t, err)
				assert.Regexp(t, tc.expectErrorMatching, err.Error())
				return
			}
			if !assert.NoError(t, err) {
				return
			}

			output := stdout.String()
			assert.Regexp(t, tc.expectOutputMatching, output)
		})
	}
}

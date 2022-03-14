package version

import (
	"bytes"
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
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
			expectErrorMatching: `expected 0 arguments, got: \[foo bar\]`,
		},
		{
			name:                 "should print version",
			thelmaArgs:           "version",
			expectOutputMatching: "version: unset\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := cli.DefaultOptions()
			options.AddCommand("version", NewVersionCommand())
			var stdout bytes.Buffer
			options.SetOut(&stdout)
			options.SetArgs(strings.Fields(tc.thelmaArgs))
			options.ConfigureThelma(func(builder builder.ThelmaBuilder) {
				builder.WithTestDefaults(t)
			})

			thelmaCLI := cli.NewWithOptions(options)
			err := thelmaCLI.Execute()
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

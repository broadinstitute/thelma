package flags

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func Test_Flags(t *testing.T) {
	type cmdOptions struct {
		dir   string
		count int
	}

	testCases := []struct {
		name                string
		input               Options
		expectedHelpMessage string
		args                []string
		expectedErr         string
		expectedOptions     cmdOptions
	}{
		{
			name: "defaults",
			input: Options{
				Prefix:      "",
				NoShortHand: false,
				Hidden:      false,
			},
			expectedHelpMessage: `
A very fake command

Usage:
  fake [flags]

Flags:
  -d, --dir string   Path to directory
      --count int    A counter
  -h, --help         help for fake
`,
			args: []string{"-d", "foo", "--count", "42"},
			expectedOptions: cmdOptions{
				dir:   "foo",
				count: 42,
			},
		},
		{
			name: "with prefix",
			input: Options{
				Prefix:      "panda",
				NoShortHand: false,
				Hidden:      false,
			},
			expectedHelpMessage: `
A very fake command

Usage:
  fake [flags]

Flags:
  -d, --panda-dir string   Path to directory
      --panda-count int    A counter
  -h, --help               help for fake
`,
			args: []string{"-d", "should/still/work", "--panda-count", "23"},
			expectedOptions: cmdOptions{
				dir:   "should/still/work",
				count: 23,
			},
		},
		{
			name: "no shorthand",
			input: Options{
				Prefix:      "",
				NoShortHand: true,
				Hidden:      false,
			},
			expectedHelpMessage: `
A very fake command

Usage:
  fake [flags]

Flags:
      --dir string   Path to directory
      --count int    A counter
  -h, --help         help for fake
`,
			args: []string{"--dir", "should/still/work", "--count", "0"},
			expectedOptions: cmdOptions{
				dir:   "should/still/work",
				count: 0,
			},
		},
		{
			name: "no shorthand should throw error",
			input: Options{
				Prefix:      "",
				NoShortHand: true,
				Hidden:      false,
			},
			expectedHelpMessage: `
A very fake command

Usage:
  fake [flags]

Flags:
      --dir string   Path to directory
      --count int    A counter
  -h, --help         help for fake
`,
			args:        []string{"-d", "should/not/work", "--count", "0"},
			expectedErr: "unknown shorthand flag: 'd'",
		},
		{
			name: "hidden",
			input: Options{
				Prefix:      "panda",
				NoShortHand: true,
				Hidden:      true,
			},
			expectedHelpMessage: `
A very fake command

Usage:
  fake [flags]

Flags:
  -h, --help         help for fake
`,
			args: []string{"--panda-dir", "should/work/again", "--panda-count", "100"},
			expectedOptions: cmdOptions{
				dir:   "should/work/again",
				count: 100,
			},
		},
	}

	for _, tc := range testCases {

		stdout := &bytes.Buffer{}
		cmd := &cobra.Command{
			Use:   "fake",
			Short: "A fake command",
			Long:  "A very fake command",
			Run: func(cmd *cobra.Command, args []string) {
				return
			},
		}
		var cmdOpts cmdOptions

		tc.input.Apply(cmd.Flags(), func(flags *pflag.FlagSet) {
			flags.StringVarP(&cmdOpts.dir, "dir", "d", "", "Path to directory")
			flags.IntVar(&cmdOpts.count, "count", 0, "A counter")
		})

		cmd.Flags().SortFlags = false
		cmd.SilenceUsage = true

		cmd.SetArgs([]string{"--help"})
		cmd.SetOut(stdout)
		require.NoError(t, cmd.Execute())
		assert.Equal(t, strings.TrimSpace(tc.expectedHelpMessage), strings.TrimSpace(stdout.String()))

		cmd.SetArgs(tc.args)
		cmd.SetOut(stdout)

		err := cmd.Execute()
		if tc.expectedErr != "" {
			require.ErrorContains(t, err, tc.expectedErr)
			return
		}

		require.NoError(t, err)
		assert.Equal(t, tc.expectedOptions.dir, cmdOpts.dir)
		assert.Equal(t, tc.expectedOptions.count, cmdOpts.count)
	}
}

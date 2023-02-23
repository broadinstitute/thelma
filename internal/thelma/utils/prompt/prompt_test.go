package prompt

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func Test_Confirm(t *testing.T) {
	testCases := []struct {
		name           string
		question       string
		defaultYes     bool
		inputLines     []string
		expectedAnswer bool
		expectedOutput string
	}{
		{
			name:       "default false",
			question:   "Do the thing?",
			defaultYes: false,
			inputLines: []string{
				"\n",
			},
			expectedAnswer: false,
			expectedOutput: `Do the thing? [y/N] `,
		},
		{
			name:       "default true",
			question:   "Do the thing?",
			defaultYes: true,
			inputLines: []string{
				"\n",
			},
			expectedAnswer: true,
			expectedOutput: `Do the thing? [Y/n] `,
		},
		{
			name:       "input matches default true",
			question:   "Do the thing?",
			defaultYes: true,
			inputLines: []string{
				"YES\n",
			},
			expectedAnswer: true,
			expectedOutput: `Do the thing? [Y/n] `,
		},
		{
			name:       "non-default false - NO",
			question:   "Do the thing?",
			defaultYes: true,
			inputLines: []string{
				"NO\n",
			},
			expectedAnswer: false,
			expectedOutput: `Do the thing? [Y/n] `,
		},
		{
			name:       "non-default true - y",
			question:   "Do the thing?",
			defaultYes: false,
			inputLines: []string{
				"y\n",
			},
			expectedAnswer: true,
			expectedOutput: `Do the thing? [y/N] `,
		},
		{
			name:       "non-default true - Y",
			question:   "Do the thing?",
			defaultYes: false,
			inputLines: []string{
				"Y\n",
			},
			expectedAnswer: true,
			expectedOutput: `Do the thing? [y/N] `,
		},
		{
			name:       "loops until valid input supplied",
			question:   "Do the thing?",
			defaultYes: false,
			inputLines: []string{
				"blah\n",
				"ye\n",
				"noooooo\n",
				"\n",
			},
			expectedAnswer: false,
			expectedOutput: `Do the thing? [y/N] Unrecognized input "blah"; please enter "y" or "n"
Do the thing? [y/N] Unrecognized input "ye"; please enter "y" or "n"
Do the thing? [y/N] Unrecognized input "noooooo"; please enter "y" or "n"
Do the thing? [y/N] `,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pipeReader, pipeWriter := io.Pipe()

			var buf bytes.Buffer
			p := &prompt{
				in:                pipeReader,
				out:               &buf,
				ensureInteractive: false,
			}

			go func() {
				for _, line := range tc.inputLines {
					_, err := pipeWriter.Write([]byte(line))
					require.NoError(t, err)
				}
				require.NoError(t, pipeWriter.Close())
			}()

			answer, err := p.Confirm(tc.question, func(options *ConfirmOptions) {
				options.DefaultYes = tc.defaultYes
			})

			require.NoError(t, err)
			assert.Equal(t, tc.expectedAnswer, answer)
			assert.Equal(t, tc.expectedOutput, buf.String())
		})
	}
}

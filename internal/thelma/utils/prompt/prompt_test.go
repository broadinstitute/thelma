package prompt

import (
	"bytes"
	"github.com/broadinstitute/thelma/internal/thelma/utils/wordwrap"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
)

func init() {
	color.NoColor = false
}

type PromptSuite struct {
	suite.Suite
	fakeStdin       *io.PipeReader
	fakeStdout      bytes.Buffer
	userInputWriter *io.PipeWriter
	prompt          Prompt
}

func TestPrompt(t *testing.T) {
	suite.Run(t, new(PromptSuite))
}

func (suite *PromptSuite) SetupSubTest() {
	pipeReader, pipeWriter := io.Pipe()

	suite.userInputWriter = pipeWriter
	suite.fakeStdin = pipeReader
	suite.fakeStdout = bytes.Buffer{}
	suite.prompt = newWith(pipeReader, &suite.fakeStdout, false, wordwrap.New())
}

func (suite *PromptSuite) TeardownSubTest() {
	require.NoError(suite.T(), suite.userInputWriter.Close())
}

func (suite *PromptSuite) Test_Print() {
	testCases := []struct {
		name     string
		input    string
		expected string
		opts     func(*PrintOptions)
	}{
		{
			name:  "should bold text if enabled",
			input: "hi",
			opts: func(options *PrintOptions) {
				options.Bold = true
			},
			expected: "\x1b[1mhi\x1b[0m\n",
		},
		{
			name:  "should pad if specified",
			input: "hi",
			opts: func(options *PrintOptions) {
				options.Bold = false
				options.LeftIndent = 6
			},
			expected: "      hi\n",
		},
		{
			name:  "should pad multiple lines if specified",
			input: "hi\nthere\nthis\nis\nfun",
			opts: func(options *PrintOptions) {
				options.Bold = false
				options.LeftIndent = 6
			},
			expected: `      hi
      there
      this
      is
      fun
`,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.prompt.Print(tc.input, tc.opts)
			require.NoError(suite.T(), err)
			output := suite.fakeStdout.String()
			assert.Equal(suite.T(), tc.expected, output)
		})
	}
}

func (suite *PromptSuite) Test_Confirm() {
	testCases := []struct {
		name           string
		question       string
		defaultYes     bool
		bold           bool
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
		{
			name:       "it should bold the prompt if enabled",
			question:   "Do the thing?",
			defaultYes: true,
			inputLines: []string{
				"\n",
			},
			bold:           true,
			expectedAnswer: true,
			expectedOutput: "\x1b[1mDo the thing?\x1b[0m [Y/n] ",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetUserInput(tc.inputLines)

			answer, err := suite.prompt.Confirm(tc.question, func(options *ConfirmOptions) {
				options.DefaultYes = tc.defaultYes
				options.Bold = tc.bold
			})

			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), tc.expectedAnswer, answer)
			assert.Equal(suite.T(), tc.expectedOutput, suite.GetOutput())
		})
	}
}

func (suite *PromptSuite) Test_Newline() {
	testCases := []struct {
		name     string
		input    []int
		expected string
	}{
		{
			name:     "no args = single newline",
			expected: "\n",
		},
		{
			name:     "1 = single newline",
			input:    []int{1},
			expected: "\n",
		},
		{
			name:     "2 = 2 newlines",
			input:    []int{2},
			expected: "\n\n",
		},
		{
			name:     "1 2 3 = 6 newlines",
			input:    []int{1, 2, 3},
			expected: "\n\n\n\n\n\n",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.prompt.Newline(tc.input...)
			require.NoError(suite.T(), err)
			output := suite.fakeStdout.String()
			assert.Equal(suite.T(), tc.expected, output)
		})
	}
}

func (suite *PromptSuite) SetUserInput(inputLines []string) {
	// save to local for safety
	t := suite.T()
	writer := suite.userInputWriter
	go func() {
		for _, line := range inputLines {
			_, err := writer.Write([]byte(line))
			require.NoError(t, err)
		}
		require.NoError(t, writer.Close())
	}()
}

func (suite *PromptSuite) GetOutput() string {
	return suite.fakeStdout.String()
}

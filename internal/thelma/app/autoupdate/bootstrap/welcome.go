package bootstrap

import (
	_ "embed"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/prompt"
	"strings"
)

//go:embed resources/ascii-logo.txt
var asciiLogo string

//go:embed resources/welcome.txt
var welcomeMessage string

// pretend terminals wider than 120 characters are still 120 chars wide
const maxTerminalWidthForPadding = 120

// welcome writes a welcome message for Thelma to the given prompt
func welcome(p prompt.Prompt) error {
	var err error

	// spacing
	if err = p.Newline(2); err != nil {
		return err
	}

	// print ascii logo
	var maxLineLen int
	for _, line := range strings.Split(asciiLogo, "\n") {
		ln := len(line)
		if ln > maxLineLen {
			maxLineLen = ln
		}
	}
	if err = p.Print(asciiLogo, func(opts *prompt.PrintOptions) {
		opts.Bold = true
		opts.LeftIndent = computeLeftPaddingToCenterLogo(utils.TerminalWidth())
	}); err != nil {
		return err
	}

	// spacing
	if err = p.Newline(2); err != nil {
		return err
	}

	// welcome message
	if err = p.Print(welcomeMessage); err != nil {
		return err
	}

	// spacing
	if err = p.Newline(2); err != nil {
		return err
	}

	return nil
}

func computeLeftPaddingToCenterLogo(terminalWidth int) int {
	// terminal width could not be detected, add a slight indent for readability
	if terminalWidth == 0 {
		return 8
	}

	if terminalWidth > maxTerminalWidthForPadding {
		terminalWidth = maxTerminalWidthForPadding
	}

	var maxLineLen int
	for _, line := range strings.Split(asciiLogo, "\n") {
		ln := len(line)
		if ln > maxLineLen {
			maxLineLen = ln
		}
	}

	// terminal is too narrow for our logo, oh well
	if maxLineLen > terminalWidth {
		return 0
	}
	return (terminalWidth - maxLineLen) / 2
}

package prompt

import (
	"bufio"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/wordwrap"
	"github.com/fatih/color"
	"io"
	"os"
	"strings"
)

type StyleOptions struct {
	Bold bool `default:"false"`
	Wrap bool `default:"true"`
}

type ConfirmOptions struct {
	DefaultYes bool `default:"true"`
	StyleOptions
}

type PrintOptions struct {
	LeftIndent int
	StyleOptions
}

type Prompt interface {
	// Confirm prompts a user for confirmation, returning true if they answer "y" or "yes"
	// and false otherwise.
	//
	// The default answer (controlled by the DefaultYes option) will be returned if
	// they hit enter without answering yes or no.
	//
	// Loops until a valid value is supplied.
	Confirm(message string, options ...func(*ConfirmOptions)) (bool, error)
	// Print is for printing a large block of text to the terminal
	Print(text string, options ...func(*PrintOptions)) error
	// Newline prints newlines to the console. If count not specified, one newline is printed
	Newline(count ...int) error
}

// New return a new Prompt instance.
func New() Prompt {
	return newWith(os.Stdin, os.Stdout, true, wordwrap.New(func(options *wordwrap.Options) {
		options.DynamicMaxWidth = true
	}))
}

// package-private constructor for testing
func newWith(in io.Reader, out io.Writer, ensureInteractive bool, wrapper wordwrap.Wrapper) Prompt {
	return &prompt{
		in:                in,
		out:               out,
		ensureInteractive: ensureInteractive,
		wrapper:           wrapper,
	}
}

type prompt struct {
	in                io.Reader
	out               io.Writer
	ensureInteractive bool
	wrapper           wordwrap.Wrapper
}

func (p *prompt) Print(text string, opts ...func(options *PrintOptions)) error {
	options := utils.CollateOptionsWithDefaults[PrintOptions](opts...)

	text = p.bold(text, options.StyleOptions)

	var maxLineLength int
	for _, line := range strings.Split(text, "\n") {
		ln := len(line)
		if ln > maxLineLength {
			maxLineLength = ln
		}
	}

	wrapLen := utils.TerminalWidth() - options.LeftIndent
	if wrapLen < 0 {
		wrapLen = 0
	}
	text = p.wrapTo(text, wrapLen, options.StyleOptions)

	leftPadding := strings.Repeat(" ", options.LeftIndent)
	for _, line := range strings.Split(text, "\n") {
		if _, err := fmt.Fprintln(p.out, leftPadding+line); err != nil {
			return err
		}
	}
	return nil
}

func (p *prompt) Confirm(message string, opts ...func(*ConfirmOptions)) (bool, error) {
	options := utils.CollateOptionsWithDefaults[ConfirmOptions](opts...)

	if err := p.verifyInteractive(); err != nil {
		return false, err
	}

	suffix := "[Y/n]"
	if !options.DefaultYes {
		suffix = "[y/N]"
	}

	message = p.bold(message, options.StyleOptions) + " " + suffix + " "

	reader := bufio.NewReader(p.in)

	for {
		if _, err := fmt.Fprint(p.out, p.wrap(message, options.StyleOptions)); err != nil {
			return false, fmt.Errorf("error prompting for user input: %v", err)
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("error prompting for user input: %v", err)
		}

		input = strings.TrimSpace(input)
		input = strings.ToLower(input)

		if len(input) == 0 { // user hit enter
			return options.DefaultYes, nil
		}
		if input == "y" || input == "yes" {
			return true, nil
		}
		if input == "n" || input == "no" {
			return false, nil
		}

		feedback := fmt.Sprintf(`Unrecognized input %q; please enter "y" or "n"`, input)
		feedback = p.wrap(feedback, options.StyleOptions)
		if _, err = fmt.Fprintln(p.out, feedback); err != nil {
			return false, fmt.Errorf("error prompting for user input: %v", err)
		}
	}
}

func (p *prompt) Newline(count ...int) error {
	var n int
	for _, c := range count {
		n += c
	}

	if n <= 0 {
		n = 1
	}

	for i := 0; i < n; i++ {
		_, err := fmt.Fprintln(p.out)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *prompt) verifyInteractive() error {
	if !p.ensureInteractive {
		return nil
	}
	if !utils.Interactive() {
		return fmt.Errorf("can't prompt for input; try re-running in an interactive shell")
	}
	return nil
}

func (p *prompt) bold(text string, option StyleOptions) string {
	if !option.Bold {
		return text
	}
	return color.New(color.Bold).Sprint(text)
}

func (p *prompt) wrap(text string, option StyleOptions) string {
	if !option.Wrap {
		return text
	}
	return p.wrapper.Wrap(text)
}

func (p *prompt) wrapTo(text string, maxWidth int, option StyleOptions) string {
	if !option.Wrap {
		return text
	}
	return p.wrapper.WrapTo(text, maxWidth)
}

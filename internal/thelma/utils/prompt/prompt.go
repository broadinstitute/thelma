package prompt

import (
	"bufio"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"io"
	"os"
	"strings"
)

type Prompt interface {
	// Confirm prompts a user for confirmation, returning true if they answer "y" or "yes"
	// and false otherwise.
	//
	// The default answer (controlled by the defaultYes parameter) will be returned if
	// they hit enter without answering yes or no.
	//
	// Loops until a valid value is supplied.
	Confirm(message string, defaultYes bool) (bool, error)
}

// New return a new Prompt instance.
func New() Prompt {
	return &prompt{
		in:                os.Stdin,
		out:               os.Stdout,
		ensureInteractive: true,
	}
}

type prompt struct {
	in                io.Reader
	out               io.Writer
	ensureInteractive bool
}

func (p *prompt) Confirm(message string, defaultYes bool) (bool, error) {
	if err := p.verifyInteractive(); err != nil {
		return false, err
	}

	suffix := "[Y/n]"
	if !defaultYes {
		suffix = "[y/N]"
	}

	reader := bufio.NewReader(p.in)

	for {
		if _, err := fmt.Fprint(p.out, "\n"+message+" "+suffix+" "); err != nil {
			return false, fmt.Errorf("error prompting for user input: %v", err)
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("error prompting for user input: %v", err)
		}

		input = strings.TrimSpace(input)
		input = strings.ToLower(input)

		if len(input) == 0 { // user hit enter
			return defaultYes, nil
		}
		if input == "y" || input == "yes" {
			return true, nil
		}
		if input == "n" || input == "no" {
			return false, nil
		}

		if _, err = fmt.Fprintf(p.out, `Unrecognized input %q; please enter "y" or "n"%s`, input, "\n"); err != nil {
			return false, fmt.Errorf("error prompting for user input: %v", err)
		}
	}
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

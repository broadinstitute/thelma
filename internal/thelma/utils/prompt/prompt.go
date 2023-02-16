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
	Confirm(message string, defaultValue bool) (bool, error)
}

// Return a new prompt instance.
func New() (Prompt, error) {
	if !utils.Interactive() {
		return nil, fmt.Errorf("can't prompt for input; try re-running in an interactive shell")
	}
	return &prompt{
		in:  os.Stdin,
		out: os.Stdout,
	}, nil
}

type prompt struct {
	in  io.Reader
	out io.Writer
}

// Confirm prompts a user for confirmation
func (p *prompt) Confirm(message string, defaultValue bool) (bool, error) {
	suffix := "[Y/n]"
	if defaultValue == false {
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
			return defaultValue, nil
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

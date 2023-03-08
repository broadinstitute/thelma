package shell

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"strings"
)

// Runner is an interface for running shell commands. It exists to
// support mocking shell commands in unit tests.
//
// https://joshrendek.com/2014/06/go-lang-mocking-exec-dot-command-using-interfaces/
type Runner interface {
	// Run runs a Command, streaming stdout and stderr to the log at debug level.
	// Behavior can be customized by passing in one or more RunOption functions
	Run(cmd Command, opts ...RunOption) error
	// PrepareSubprocess sets up a Subprocess to run a Command, similar to Run but asynchronous.
	PrepareSubprocess(cmd Command, opts ...RunOption) Subprocess
}

// RunOptions are option for a RunWith() invocation
type RunOptions struct {
	// optional logger to use for logging this command
	Logger *zerolog.Logger
	// optional level at which command should be logged
	LogLevel zerolog.Level
	// optional level at which command output (stdout/stderr) should be logged
	OutputLogLevel zerolog.Level
	// optional reader from which stdin should be read
	Stdin io.Reader
	// optional writer where stdout should be written
	Stdout io.Writer
	// optional writer where stderr should be written
	Stderr io.Writer
	// if false, do not send stdout to logging system
	LogStdout bool
}

// RunOption can be used to configure RunOptions for a Run invocation
type RunOption func(*RunOptions)

// Command encapsulates a shell command
type Command struct {
	Prog        string   // Prog Main CLI program to execute
	Args        []string // Args Arguments to pass to program
	Env         []string // Env List of environment variables, eg []string{ "FOO=BAR", "BAZ=QUUX" }, to set when executing
	Dir         string   // Dir Directory where command should be run
	PristineEnv bool     // PristineEnv When true, set only supplied Env vars without inheriting current process's env vars
}

// Subprocess allows running a shell command in parallel with Thelma's main process
type Subprocess interface {
	// Start begins running the actual Command the Subprocess is configured for
	Start() error
	// Wait synchronously blocks until normal completion of the Command
	Wait() error
	// Stop signals the process to exit, and synchronously waits until it does; after three seconds it will be forcibly
	// killed and an error returned
	Stop() error
}

func defaultRunOptions() RunOptions {
	return RunOptions{
		LogLevel:       zerolog.DebugLevel,
		OutputLogLevel: zerolog.DebugLevel,
		LogStdout:      true,
	}
}

// PrettyFormat converts a command into a simple string for easy inspection. Eg.
//
//	&Command{
//	  Prog: []string{"echo"},
//	  Args: []string{"foo", "bar", "baz"},
//	  Dir:  "/tmp",
//	  Env:  []string{"A=B", "C=D"}
//	}
//
// ->
// "A=B C=D echo foo bar baz"
func (c Command) PrettyFormat() string {
	// TODO shellquote arguments for better readability
	var a []string
	a = append(a, c.Env...)
	a = append(a, c.Prog)
	a = append(a, c.Args...)
	return strings.Join(a, " ")
}

// Error is a generic error that is returned in situations other than the command failing.
// (eg. if the Command's Directory does not exist)
type Error struct {
	Command Command // the command that generated this error e
	err     error   // underlying error returned by exec package
}

// Error generates a user-friendly error message
func (e *Error) Error() string {
	cmd := e.Command.PrettyFormat()
	return fmt.Sprintf("Command %q failed to start: %v", cmd, e.err)
}

// ExitError is returned when a command fails
type ExitError struct {
	Command  Command // the command that generated this error
	ExitCode int     // exit code of command
	Stderr   string  // stderr output
}

// Error generates a user-friendly error message for failed shell commands
func (e *ExitError) Error() string {
	cmd := e.Command.PrettyFormat()
	msg := fmt.Sprintf("Command %q exited with status %d", cmd, e.ExitCode)
	stderr := e.Stderr
	// Add stderr output if any was generated
	if len(stderr) > 0 {
		msg = fmt.Sprintf("%s:\n%s", msg, stderr)
	}

	return msg
}

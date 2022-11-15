package shell

import (
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/logid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"time"
)

// RealRunner is an implementation of the Runner interface that actually executes shell commands
type RealRunner struct{}

// NewRunner constructs a new Runner
func NewRunner() Runner {
	return &RealRunner{}
}

// Run runs a Command, returning an error if the command exits non-zero
func (r *RealRunner) Run(cmd Command, options ...RunOption) error {
	execCmd, logger, errBuf := r.prepareExecCmd(cmd, options...)
	err := execCmd.Run()
	return handleExecCmdError(cmd, err, logger, errBuf)
}

// PrepareSubprocess sets up a Subprocess to run a Command asynchronously
func (r *RealRunner) PrepareSubprocess(cmd Command, options ...RunOption) Subprocess {
	execCmd, logger, errCapture := r.prepareExecCmd(cmd, options...)
	return &realSubprocess{
		cmd:       cmd,
		execCmd:   execCmd,
		logger:    logger,
		errWriter: errCapture,
	}
}

func (r *RealRunner) prepareExecCmd(cmd Command, options ...RunOption) (*exec.Cmd, zerolog.Logger, *bytes.Buffer) {
	// collate options
	opts := defaultRunOptions()
	for _, option := range options {
		option(&opts)
	}

	// handle options
	logger := log.Logger
	if opts.Logger != nil {
		logger = *opts.Logger
	}
	level := opts.LogLevel
	stderr := opts.Stderr
	stdout := opts.Stdout

	// Generate an id to uniquely identify this command in log messages and add to Log context
	logger = logger.With().Str("cmd", logid.NewId()).Logger()

	// Wrap user-supplied stderr writer in a new io.Writer that records stderr output
	errBuffer := &bytes.Buffer{}
	writers := []io.Writer{errBuffer}
	if stderr != nil {
		writers = append(writers, stderr)
	}
	errWriter := io.MultiWriter(writers...)

	// Wrap user-supplied stdout and stderr in new io.Writers that log messages at debug level
	if opts.LogStdout {
		stdout = NewLoggingWriter(opts.OutputLogLevel, logger.With().Str("stream", "stdout").Logger(), "[out] ", stdout)
	}
	stderr = NewLoggingWriter(opts.OutputLogLevel, logger.With().Str("stream", "stderr").Logger(), "[err] ", errWriter)

	// Convert our command arguments to exec.Cmd struct
	execCmd := exec.Command(cmd.Prog, cmd.Args...)
	execCmd.Dir = cmd.Dir
	if !cmd.PristineEnv {
		execCmd.Env = os.Environ()
	}
	execCmd.Env = append(execCmd.Env, cmd.Env...)
	execCmd.Stdout = stdout
	execCmd.Stderr = stderr

	logger.WithLevel(level).Str("dir", cmd.Dir).Msgf("Executing: %q", cmd.PrettyFormat())
	return execCmd, logger, errBuffer
}

func handleExecCmdError(cmd Command, err error, logger zerolog.Logger, errBuf *bytes.Buffer) error {
	if err != nil {
		logger.Debug().Msgf("Command failed: %v\n", err)

		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{
				Command:  cmd,
				ExitCode: exitErr.ExitCode(),
				Stderr:   errBuf.String(),
			}
		} else {
			return &Error{
				Command: cmd,
				err:     err,
			}
		}
	}
	return nil
}

type realSubprocess struct {
	cmd       Command
	execCmd   *exec.Cmd
	logger    zerolog.Logger
	errWriter io.Writer
	errBuf    *bytes.Buffer
}

func (s *realSubprocess) Start() error {
	return s.execCmd.Start()
}

func (s *realSubprocess) Wait() error {
	return handleExecCmdError(s.cmd, s.execCmd.Wait(), s.logger, s.errBuf)
}

func (s *realSubprocess) Stop() error {
	if s.execCmd.ProcessState != nil && s.execCmd.ProcessState.Exited() {
		log.Debug().Msg("process had already exited")
		return nil
	}
	if s.execCmd.Process == nil {
		return fmt.Errorf("no process associated with command")
	} else {
		if err := s.execCmd.Process.Signal(os.Interrupt); err != nil {
			// Can't send SIGINT on Windows; it'll error, so send SIGKILL
			if err := s.execCmd.Process.Signal(os.Kill); err != nil {
				// If signals fail, just kill the underlying process
				if err := s.execCmd.Process.Kill(); err != nil {
					log.Debug().Msg("seemed to be unable to SIGINT, SIGKILL, or directly kill a process...")
				}
			}
		}
		done := make(chan error)
		go func() {
			done <- s.execCmd.Wait()
		}()
		select {
		case err := <-done:
			return handleExecCmdError(s.cmd, err, s.logger, s.errBuf)
		case <-time.After(3 * time.Second):
			_ = s.execCmd.Process.Kill()
			return fmt.Errorf("process did not exit after 3 seconds")
		}
	}
}

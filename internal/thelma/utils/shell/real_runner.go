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
)

const maxErrorBufLenBytes = 100 * 1024 // 100 kb
const eol = '\n'

// RealRunner is an implementation of the Runner interface that actually executes shell commands
type RealRunner struct{}

// NewRunner constructs a new Runner
func NewRunner() Runner {
	return &RealRunner{}
}

// Run runs a Command, returning an error if the command exits non-zero
func (r *RealRunner) Run(cmd Command, options ...RunOption) error {
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
	errCapture := newCapturingWriter(maxErrorBufLenBytes, logger, stderr)

	// Wrap user-supplied stdout and stderr in new io.Writers that log messages at debug level
	stdout = NewLoggingWriter(opts.OutputLogLevel, logger.With().Str("stream", "stdout").Logger(), "[out] ", stdout)
	stderr = NewLoggingWriter(opts.OutputLogLevel, logger.With().Str("stream", "stderr").Logger(), "[err] ", errCapture)

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

	err := execCmd.Run()
	if err != nil {
		logger.Debug().Msgf("Command failed: %v\n", err)

		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{
				Command:  cmd,
				ExitCode: exitErr.ExitCode(),
				Stderr:   errCapture.String(),
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

// An io.Writer that captures data it receives with Write() into a buffer and optionally forwards to another writer
type capturingWriter struct {
	len    int
	maxLen int
	buf    *bytes.Buffer
	logger zerolog.Logger
	inner  io.Writer
}

func newCapturingWriter(rolloverLen int, logger zerolog.Logger, inner io.Writer) *capturingWriter {
	return &capturingWriter{
		maxLen: rolloverLen,
		buf:    bytes.NewBuffer([]byte{}),
		inner:  inner,
		logger: logger,
	}
}

func (cw *capturingWriter) String() string {
	return cw.buf.String()
}

func (cw *capturingWriter) Write(p []byte) (n int, err error) {
	if len(p) > cw.maxLen {
		if cw.len > 0 {
			cw.rollover()
		}

		cw.logger.Warn().
			Int("max-len", cw.maxLen).
			Str("content", string(p)).
			Msgf("capturing writer: message too long (%d bytes), won't capture", len(p))

		n, err = len(p), nil
	} else {
		if cw.len+len(p) > cw.maxLen {
			cw.rollover()
		}

		n, err = cw.buf.Write(p)
		if err != nil {
			return n, fmt.Errorf("capturing writer: error writing to buffer: %v", err)
		}
		cw.len += n
	}

	if cw.inner == nil {
		return n, err
	}

	return cw.inner.Write(p)
}

func (cw *capturingWriter) rollover() {
	cw.logger.Warn().
		Int("current-len", cw.len).
		Int("max-len", cw.maxLen).
		Str("content", cw.buf.String()).
		Msg("capturing writer: buffer rolled over")
	cw.buf = bytes.NewBuffer([]byte{})
	cw.len = 0
}

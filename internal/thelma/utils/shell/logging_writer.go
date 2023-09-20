package shell

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"strings"
)

const eol = '\n'

// LoggingWriter an io.Writer that logs messages that are sent to it with Write() and optionally forwards to another io.Writer
type LoggingWriter struct {
	level  zerolog.Level
	logger zerolog.Logger
	prefix string
	inner  io.Writer
}

func NewLoggingWriter(level zerolog.Level, logger zerolog.Logger, prefix string, inner io.Writer) *LoggingWriter {
	return &LoggingWriter{
		level:  level,
		logger: logger,
		prefix: prefix,
		inner:  inner,
	}
}

func (lw *LoggingWriter) Write(p []byte) (n int, err error) {
	n, err = lw.streamLinesToLog(p)

	if lw.inner == nil {
		return n, err
	}

	return lw.inner.Write(p)
}

func (lw *LoggingWriter) streamLinesToLog(p []byte) (n int, err error) {
	p2 := make([]byte, len(p))
	copy(p2, p)

	eolStr := string(eol)

	buf := bytes.NewBuffer(p2)
	for {
		line, err := buf.ReadString(eol)
		n += len(line)

		if err == nil || len(line) > 0 {
			line = strings.TrimSuffix(line, eolStr)
			lw.logger.WithLevel(lw.level).Msgf("%s%s", lw.prefix, line)
		}

		if err == io.EOF {
			return n, nil
		} else if err != nil {
			return n, errors.Errorf("logging writer: error reading from buffer: %v", err)
		}
	}
}

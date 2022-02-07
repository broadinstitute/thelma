package logging

import (
	"github.com/rs/zerolog"
)

// Implement a custom LevelWriter to support logging to multiple destinations
// as suggested here: https://github.com/rs/zerolog/issues/150

// FilteredWriter drops log messages below a specified threshold
type FilteredWriter struct {
	// inner writer to send log messages to
	inner zerolog.LevelWriter
	// log messages < this level are ignored
	filterBelow zerolog.Level
}

func NewFilteredWriter(inner zerolog.LevelWriter, filterBelow zerolog.Level) zerolog.LevelWriter {
	return &FilteredWriter{
		inner:       inner,
		filterBelow: filterBelow,
	}
}

func (w *FilteredWriter) Write(p []byte) (n int, err error) {
	return w.inner.Write(p)
}

func (w *FilteredWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= w.filterBelow {
		return w.inner.Write(p)
	}
	return len(p), nil
}

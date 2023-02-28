package logging

import (
	"github.com/broadinstitute/thelma/internal/thelma/utils/wordwrap"
	"io"
)

// WrappingWriter wraps lines in log messages to a given width
type WrappingWriter struct {
	// inner writer to send log messages to
	inner io.Writer
	// wordwrapper instance used to wrap lines
	wrapper wordwrap.Wrapper
}

func NewWrappingWriter(inner io.Writer, opts ...func(*wordwrap.Options)) io.Writer {
	return &WrappingWriter{
		inner:   inner,
		wrapper: wordwrap.New(opts...),
	}
}

func (w *WrappingWriter) Write(p []byte) (n int, err error) {
	wrapped := []byte(w.wrapper.Wrap(string(p)))
	_, err = w.inner.Write(wrapped)

	// note that we return len(p) because it confuses upstream
	// writers if we return a "bytes written" count higher than the length of
	// p, the input they passed us.
	// more context here: https://groups.google.com/g/golang-nuts/c/HGDnA7-gXN4
	return len(p), err
}

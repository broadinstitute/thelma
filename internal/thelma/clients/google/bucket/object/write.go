package object

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
)

type Write interface {
	Operation
}

func NewWrite(content []byte, attrs AttrSet) Write {
	return &write{
		content: content,
		attrs:   attrs,
	}
}

func NewWriteFromStream(reader io.Reader, attrs AttrSet) Write {
	return &write{
		reader: reader,
		attrs:  attrs,
	}
}

type write struct {
	content []byte
	reader  io.Reader
	attrs   AttrSet
}

func (w *write) Kind() string {
	return "write"
}

func (w *write) Handler(object Object, logger zerolog.Logger) error {
	writer := object.Handle.NewWriter(object.Ctx)
	w.attrs.writeToLogEvent(logger.Debug())
	w.attrs.applyToWriter(writer)

	written, err := w.writeContent(writer)
	if err != nil {
		return fmt.Errorf("error writing object: %v", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("error closing writer: %v", err)
	}

	logTransfer(logger, int64(written))
	return nil
}

func (w *write) writeContent(writer io.Writer) (int64, error) {
	if w.reader != nil {
		return io.Copy(writer, w.reader)
	} else {
		written, err := writer.Write(w.content)
		return int64(written), err
	}
}

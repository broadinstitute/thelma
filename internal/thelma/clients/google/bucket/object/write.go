package object

import (
	"fmt"
	"github.com/rs/zerolog"
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

type write struct {
	content []byte
	attrs   AttrSet
}

func (w *write) Kind() string {
	return "write"
}

func (w *write) Handler(object Object, logger zerolog.Logger) error {
	writer := object.Handle.NewWriter(object.Ctx)
	w.attrs.writeToLogEvent(logger.Debug())
	w.attrs.applyToWriter(writer)

	written, err := writer.Write(w.content)
	if err != nil {
		return fmt.Errorf("error writing object: %v", err)
	}
	if err = writer.Close(); err != nil {
		return fmt.Errorf("error closing writer: %v", err)
	}

	logTransfer(logger, int64(written))
	return nil
}

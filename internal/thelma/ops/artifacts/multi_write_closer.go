package artifacts

import (
	"io"
)

func newMultiWriteCloser(writeClosers ...io.WriteCloser) *multiWriteCloser {
	var writers []io.Writer
	for _, writeCloser := range writeClosers {
		writers = append(writers, writeCloser)
	}

	return &multiWriteCloser{
		multiWriter:  io.MultiWriter(writers...),
		writeClosers: writeClosers,
	}
}

// simple multiWriteCloser. Like io.MultiWriter, but it also supports Close()
type multiWriteCloser struct {
	multiWriter  io.Writer
	writeClosers []io.WriteCloser
}

func (m *multiWriteCloser) Write(p []byte) (n int, err error) {
	return m.multiWriter.Write(p)
}

func (m *multiWriteCloser) Close() error {
	for _, writeCloser := range m.writeClosers {
		if err := writeCloser.Close(); err != nil {
			return err
		}
	}
	return nil
}

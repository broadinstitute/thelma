package bucket

import (
	"io"
)

func newLoggingWriteCloser(objectWriteCloser io.WriteCloser, opLogger *operationLogger) io.WriteCloser {
	return &loggingWriteCloser{
		objectWriteCloser: objectWriteCloser,
		opLogger:          opLogger,
	}
}

type loggingWriteCloser struct {
	objectWriteCloser io.WriteCloser
	opLogger          *operationLogger
	started           bool
}

func (l *loggingWriteCloser) Write(p []byte) (n int, err error) {
	l.logStartIfNeeded()
	return l.objectWriteCloser.Write(p)
}

func (l *loggingWriteCloser) logStartIfNeeded() {
	if l.started {
		return
	}
	l.started = true
	l.opLogger.operationStarted()
}

func (l *loggingWriteCloser) Close() error {
	err := l.objectWriteCloser.Close()
	l.opLogger.operationFinished(err)
	return err
}

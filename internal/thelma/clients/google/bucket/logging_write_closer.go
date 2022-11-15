package bucket

import (
	"io"
)

// loggingWriteCloser: given a google cloud storage object write closer (for which Write() writes data to a GCS object),
// wrap it in a new WriteCloser that logs operation start and stop on the first Write() and first Close() calls,
// respectively
func newLoggingWriteCloser(objectWriteCloser io.WriteCloser, opLogger *operationLogger) io.WriteCloser {
	return &loggingWriteCloser{
		objectWriteCloser: objectWriteCloser,
		opLogger:          opLogger,
	}
}

type loggingWriteCloser struct {
	objectWriteCloser io.WriteCloser
	opLogger          *operationLogger
}

func (l *loggingWriteCloser) Write(p []byte) (n int, err error) {
	l.opLogger.operationStarted()
	return l.objectWriteCloser.Write(p)
}

func (l *loggingWriteCloser) Close() error {
	err := l.objectWriteCloser.Close()
	err = l.opLogger.operationFinished(err)
	return err
}

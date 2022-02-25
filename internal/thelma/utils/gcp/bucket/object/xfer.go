package object

import "github.com/rs/zerolog"

// utility methods for operations that transfer object content (upload, download, read, write)
const logFieldContentLength = "content-length"
const logFieldFile = "file"

func logTransfer(logger zerolog.Logger, contentLength int64) {
	logger.Debug().Int64(logFieldContentLength, contentLength).Msgf("transferred %d bytes", contentLength)
}

func addFileCtx(logger zerolog.Logger, file string) zerolog.Logger {
	return logger.With().Str(logFieldFile, file).Logger()
}

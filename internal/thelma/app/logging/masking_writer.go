package logging

import (
	"bytes"
	"github.com/rs/zerolog"
	"io"
)

const replacementText = "******"

// MaskingWriter replaces sensitive text in log messages
type MaskingWriter struct {
	// inner writer to send log messages to
	inner io.Writer
	// secrets list of strings to redact from log messages (as byte slices)
	secrets [][]byte
	// replacementText text to replace secrets with (as byte slice)
	replacementText []byte
}

func NewMaskingWriter(inner io.Writer, secrets []string) zerolog.LevelWriter {
	// convert secrets to byte array
	_secrets := make([][]byte, len(secrets))
	for i, s := range secrets {
		_secrets[i] = []byte(s)
	}

	return &MaskingWriter{
		inner:           inner,
		secrets:         _secrets,
		replacementText: []byte(replacementText),
	}
}

// Note that Zerolog guarantees log messages won't be split up across multiple Write calls:
// "Each logging operation makes a single call to the Writer's Write method."
// (https://github.com/rs/zerolog/blob/e9344a8c507b5f25a4962ff022526be0ddab8e72/log.go#L210)
// So we don't need to worry about secrets potentially being split across multiple writes.
func (w *MaskingWriter) Write(p []byte) (n int, err error) {
	return w.inner.Write(w.redactSecrets(p))
}

func (w *MaskingWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	p = w.redactSecrets(p)

	asLevelWriter, ok := w.inner.(zerolog.LevelWriter)
	if !ok {
		return w.inner.Write(p)
	}
	return asLevelWriter.WriteLevel(level, p)
}

func (w *MaskingWriter) redactSecrets(msg []byte) []byte {
	for _, secret := range w.secrets {
		msg = bytes.ReplaceAll(msg, secret, w.replacementText)
	}
	return msg
}

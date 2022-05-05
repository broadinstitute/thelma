package logging

import (
	"bytes"
	"github.com/rs/zerolog"
	"io"
	"sync"
)

const replacementText = "******"

// MaskingWriter replaces sensitive text in log messages
type MaskingWriter struct {
	// inner writer to forward log messages to
	inner io.Writer
	// secrets set of strings to redact from log messages (kept as byte slices keyed by string representation)
	secrets map[string][]byte
	// replacementText text to replace secrets with (as byte slice)
	replacementText []byte
	// mutex used to protect updates to the secrets list
	mutex sync.RWMutex
}

func NewMaskingWriter(inner io.Writer) *MaskingWriter {
	return &MaskingWriter{
		inner:           inner,
		replacementText: []byte(replacementText),
		secrets:         make(map[string][]byte),
	}
}

func (w *MaskingWriter) MaskSecrets(secrets ...string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// convert secrets to byte arrays
	for _, s := range secrets {
		_, exists := w.secrets[s]
		if !exists {
			w.secrets[s] = []byte(s)
		}
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
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	for _, secret := range w.secrets {
		msg = bytes.ReplaceAll(msg, secret, w.replacementText)
	}
	return msg
}

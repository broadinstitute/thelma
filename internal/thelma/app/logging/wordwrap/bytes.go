package wordwrap

import (
	"bytes"
	"unicode/utf8"
)

type BytesWrapper wordWrapper[[]byte]

func ForBytes(opts ...func(*Options)) BytesWrapper {
	return newGenericWrapper(bytesDecoder, newBytesBuffer, opts...)
}

func bytesDecoder(v []byte, i int) (rune, int) {
	return utf8.DecodeRune(v[i:])
}

func newBytesBuffer() buffer[[]byte] {
	return &bytesBuffer{}
}

type bytesBuffer struct {
	bytes.Buffer
}

func (b *bytesBuffer) Get() []byte {
	return b.Bytes()
}

package wordwrap

import (
	"strings"
	"unicode/utf8"
)

type StringWrapper wordWrapper[string]

func ForString(opts ...func(*Options)) StringWrapper {
	return newGenericWrapper(stringDecoder, newStringBuffer, opts...)
}

func stringDecoder(v string, i int) (rune, int) {
	return utf8.DecodeRuneInString(v[i:])
}

func newStringBuffer() buffer[string] {
	return &stringBuffer{}
}

type stringBuffer struct {
	strings.Builder
}

func (b *stringBuffer) Get() string {
	return b.String()
}

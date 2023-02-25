package wordwrap

type wrappable interface {
	[]byte | string
}

type decoder[T wrappable] func(v T, i int) (rune, int)

type buffer[T wrappable] interface {
	Write(p []byte) (n int, err error)
	WriteByte(b byte) error
	WriteString(s string) (n int, err error)
	WriteRune(r rune) (n int, err error)
	Get() T
}

package logging

import (
	"fmt"
	"github.com/rs/zerolog"
	"strings"
	"unicode"
)

// WrappingWriter drops log messages below a specified threshold
type WrappingWriter struct {
	// inner writer to send log messages to
	inner zerolog.LevelWriter
	opts  WrapOptions
}

func NewWrappingWriter(inner zerolog.LevelWriter, opts WrapOptions) zerolog.LevelWriter {
	return &WrappingWriter{
		inner: inner,
		opts:  opts,
	}
}

func (w *WrappingWriter) Write(p []byte) (n int, err error) {
	wrapped := newWordWrapper(string(p), w.opts).wrap()
	return w.inner.Write([]byte(wrapped))
}

func (w *WrappingWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	wrapped := newWordWrapper(string(p), w.opts).wrap()
	return w.inner.WriteLevel(level, []byte(wrapped))
}

func wrap(input string, maxWidth int, padding string, escapeStringLiteral bool) string {
	w := newWordWrapper(input, WrapOptions{
		maxWidth:                   maxWidth,
		padding:                    padding,
		escapeNewlineStringLiteral: escapeStringLiteral,
	})

	return w.wrap()
}

type wordWrapper struct {
	opts       WrapOptions      // wrapping options
	s          []rune           // the original, unwrapped string converted to a rune slice
	i          int              // in s: start of current word, and the first character not yet copied to buf
	n          int              // in s: start of the next word
	quoteCount int              // tracks number of double-quote characters ('"') we've seen so far in s
	buf        *strings.Builder // buf stores the word-wrapped string we're building as we traverse s
	lineOffset int              // in buf: number of characters written since last line break
}

// okay so one way to walk through the string would be to not index on the rune.
func newWordWrapper(input string, opts WrapOptions) *wordWrapper {
	return &wordWrapper{
		opts:       opts,
		s:          []rune(input),
		i:          0,
		n:          0,
		quoteCount: 0,
		buf:        &strings.Builder{},
		lineOffset: 0,
	}
}

type WrapOptions struct {
	maxWidth                   int
	padding                    string
	escapeNewlineStringLiteral bool
}

func (w *wordWrapper) wrap() string {
	for w.nextWord() {
		fmt.Printf("I: %d N: %d\n", w.i, w.n)
		if w.wordExceedsMaxWidth() && !w.atStartOfLine() {
			w.startNewLine()
		}
		w.copyWordToBuffer()
	}

	return w.buf.String()
}

func (w *wordWrapper) copyWordToBuffer() {
	for k := w.i; k < w.n; k++ {
		w.buf.WriteRune(w.s[k])
	}
	w.lineOffset += w.wordLength()
}

func (w *wordWrapper) startNewLine() {
	if w.opts.escapeNewlineStringLiteral && w.betweenQuoteMarks() {
		// add ` \` to end of current line before adding a new line
		w.buf.WriteString(" \\")
	}

	// if the first character of the current word is whitespace, "replace" it
	// with the newline by incrementing i before the word is copied.
	// This way wrapped lines will not start with a leading space.
	if unicode.IsSpace(w.s[w.i]) {
		w.i++
	}

	// write newline and reset line offset
	w.buf.WriteByte('\n')
	w.lineOffset = 0

	// write optional padding after newline
	w.buf.WriteString(w.opts.padding)
	w.lineOffset += len(w.opts.padding)
}

func (w *wordWrapper) wordExceedsMaxWidth() bool {
	maxWidth := w.opts.maxWidth
	if w.opts.escapeNewlineStringLiteral && w.betweenQuoteMarks() {
		// take off 2 chars to leave room for escape sequence
		maxWidth = w.opts.maxWidth - 2
	}
	return (w.lineOffset + w.wordLength()) > maxWidth
}

func (w *wordWrapper) wordLength() int {
	return w.n - w.i
}

func (w *wordWrapper) atStartOfLine() bool {
	return w.lineOffset == 0
}

func (w *wordWrapper) finished() bool {
	// last character in s has been copied to buf
	return w.i == len(w.s)
}

func (w *wordWrapper) betweenQuoteMarks() bool {
	return w.quoteCount%2 == 1
}

// move on to the next word, returning false if we've reached end of s
func (w *wordWrapper) nextWord() bool {
	w.i = w.n
	w.n = findStartOfNextWord(w.s, w.n)

	w.quoteCount += countQuotes(w.s, w.i, w.n)

	return w.i != w.n
}

func countQuotes(s []rune, startAt int, endBefore int) int {
	count := 0
	for i := startAt + 1; i < endBefore; i++ {
		if s[i] == '"' {
			count++
		}
	}
	return count
}

func findStartOfNextWord(s []rune, startAt int) int {
	// ignore leading whitespace -- counts as part of the word
	i := startAt
	for i < len(s) && unicode.IsSpace(s[i]) {
		i++
	}

	// we reached a non-whitespace character. read all the non-whitespace characters until we
	// get to another whitespace character.
	for i < len(s) && !unicode.IsSpace(s[i]) {
		i++
	}
	return i
}

package wordwrap

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh/terminal"
	"unicode"
)

const doubleQuote = '"'
const newLine = '\n'
const space = ' '
const backslash = '\\'

// quoteState is used to track whether we're inside double quotes, for
// the EscapeNewlineStringLiteral option
type quoteState int64

const (
	outside quoteState = iota // outside double quotes
	starts                    // this word start a double quote, eg. `"foo`
	inside                    // inside double quotes
	ends                      // this words ends a double quote, eg. `end"`
)

// Options options for a word wrapper
type Options struct {
	MaxWidth                   int    // MaxWidth maximum width of line / length at which newline should be inserted
	Padding                    string // Padding string to pad wrapped lines with. Eg. "  " for 2-space indent
	EscapeNewlineStringLiteral bool   // EscapeNewlineStringLiteral For `"long strings like this"`, indicates whether inserted newlines should be prefixed with \"
}

type wordWrapper[T wrappable] interface {
	Wrap(input T) (output T)
}

func newGenericWrapper[T wrappable](decoder decoder[T], allocator func() buffer[T], opts ...func(*Options)) wordWrapper[T] {
	options := utils.CollateOptions(Options{
		MaxWidth:                   defaultMaxWidth(),
		Padding:                    "",
		EscapeNewlineStringLiteral: false,
	}, opts...)

	var padOffset int
	for _, r := range options.Padding {
		if unicode.IsPrint(r) {
			padOffset++
		}
	}

	return wrapper[T]{
		options:   options,
		allocator: allocator,
		decoder:   decoder,
		padOffset: padOffset,
	}
}

// wrapper implements the wordWrapper interface
type wrapper[T wrappable] struct {
	options   Options
	allocator func() buffer[T] // allocates new buffers for building output
	decoder   decoder[T]       // decodes runes from input
	padOffset int              // number of printable characters in options.Padding

}

func (w wrapper[T]) Wrap(s T) T {
	if w.options.MaxWidth <= 0 {
		return s
	}
	return newLineBreaker(s, w).wrap()
}

type linebreaker[T wrappable] struct {
	opts       Options    // wrapping options
	s          T          // the original, unwrapped string or []byte
	decoder    decoder[T] // decodes runes from the input type (string or []byte)
	buf        buffer[T]  // stores the word-wrapped string we're building as we traverse s
	i          int        // in s: start of current word, and the first byte not yet copied to buf
	n          int        // in s: start of the next word
	quoteState quoteState // tracks whether current word is inside unescaped quotation marks
	lineOffset int        // in buf: number of characters written since last line break
	padOffset  int        // number of printable characters in Padding
}

func newLineBreaker[T wrappable](input T, w wrapper[T]) *linebreaker[T] {
	return &linebreaker[T]{
		opts:       w.options,
		decoder:    w.decoder,
		s:          input,
		i:          0,
		n:          0,
		buf:        w.allocator(),
		lineOffset: 0,
		padOffset:  w.padOffset,
	}
}

func (b *linebreaker[T]) wrap() T {
	for b.advanceToNextWord() {
		if b.wordExceedsMaxWidth() && !b.atStartOfLine() {
			b.startNewLine()
		}
		b.copyWordToBuffer()
	}

	return b.buf.Get()
}

// copy current word to buffer, incrementing line offset as we go
func (b *linebreaker[T]) copyWordToBuffer() {
	for j := b.i; j < b.n; {
		r, ln := b.runeAt(j)
		_, _ = b.buf.WriteRune(r)

		if r == newLine {
			b.lineOffset = 0
		} else if unicode.IsPrint(r) {
			// only increment line offset if r is printable - so that "invisible" runes like
			// ansii color escape sequences don't count towards the line offset
			b.lineOffset++
		} else {
			fmt.Printf("ignored unprintable character: %+q\n", r)
		}

		j += ln
	}
}

func (b *linebreaker[T]) startNewLine() {
	fmt.Printf("start newline\n")
	if b.shouldEscapeNewlineBeforeWrapping() {
		// add ` \` to end of current line before adding a new line
		_ = b.buf.WriteByte(space)
		_ = b.buf.WriteByte(backslash)
	}

	// if the first character of the current word is whitespace, "replace" it
	// with the newline by incrementing i before the word is copied.
	// This way wrapped lines will not start with a leading space.
	if r, _ := b.runeAt(b.i); unicode.IsSpace(r) {
		b.i++
	}

	// write newline and reset line offset
	_ = b.buf.WriteByte(newLine)
	b.lineOffset = 0

	// write optional padding after newline
	_, _ = b.buf.WriteString(b.opts.Padding)
	b.lineOffset += b.padOffset
}

func (b *linebreaker[T]) wordExceedsMaxWidth() bool {
	maxWidth := b.opts.MaxWidth
	if b.shouldEscapeNewlineBeforeWrapping() {
		// take off 2 chars to leave room for escape sequence
		maxWidth = b.opts.MaxWidth - 2
	}
	// if this is to avoid allocating memory, like, meh.
	//
	// okay so we have the output buffer and a word buffer.
	// as we read through the
	//
	wordLen := b.wordLength()
	fmt.Printf("i: %d, n: %d\n", b.i, b.n)
	fmt.Printf("offset: %d\n", b.lineOffset)
	fmt.Printf("word len: %d\n", wordLen)
	fmt.Printf("max wid: %d\n", maxWidth)
	return (b.lineOffset + b.wordLength()) > maxWidth
}

func (b *linebreaker[T]) wordLength() int {
	var count int
	for j := b.i; j < b.n; {
		r, ln := b.runeAt(j)
		fmt.Printf("word len: %c\n", r)

		if unicode.IsPrint(r) {
			// ignore unprintable characters when computing word length
			count++
		}
		j += ln
	}
	return count
}

func (b *linebreaker[T]) atStartOfLine() bool {
	return b.lineOffset == 0
}

func (b *linebreaker[T]) finished() bool {
	// last character in s has been copied to buf
	return b.i == len(b.s)
}

// move on to the next word, returning false if we've reached end of s
func (b *linebreaker[T]) advanceToNextWord() bool {
	b.i = b.n
	b.n = b.findStartOfNextWord(b.n)

	b.countQuotes()

	return b.i != b.n
}

func (b *linebreaker[T]) shouldEscapeNewlineBeforeWrapping() bool {
	if !b.opts.EscapeNewlineStringLiteral {
		return false
	}
	return b.quoteState == inside || b.quoteState == ends
}

// scan through substring and check for quote marks
func (b *linebreaker[T]) countQuotes() {
	var count int
	var prev rune

	// iterate through runes in current word
	for j := b.i; j < b.n; {
		r, ln := b.runeAt(j)
		if r == doubleQuote {
			// make sure previous rune was not a \
			if j == b.i || prev != backslash {
				count++
			}
		}
		prev = r
		j += ln
	}

	b.quoteState = nextQuoteState(b.quoteState, count)
}

func nextQuoteState(previous quoteState, quoteCount int) quoteState {
	if quoteCount%2 == 1 {
		// odd number of quotes in current word -- this word either starts a quote or ends it
		if previous == starts || previous == inside {
			return ends
		} else if previous == ends || previous == outside {
			return starts
		} else {
			panic(fmt.Errorf("unmatched quote state: %v", previous))
		}
	} else {
		// this word is not a quote boundary - if we were previously at a boundary, transition to outside/inside
		if previous == starts {
			return inside
		} else if previous == ends {
			return outside
		} else {
			return previous
		}
	}
}

func (b *linebreaker[T]) findStartOfNextWord(startAt int) int {
	j := startAt

	// ignore leading whitespace -- counts as part of the word
	j = b.firstRuneMatching(j, utils.Not(unicode.IsSpace))

	// we reached a non-whitespace character. Now read all the non-whitespace characters until we
	// get to another whitespace character -- the start of the next word (or end of input)
	j = b.firstRuneMatching(j, unicode.IsSpace)

	return j
}

func (b *linebreaker[T]) firstRuneMatching(startAt int, predicate func(rune) bool) int {
	j := startAt
	for j < len(b.s) {
		r, rlen := b.runeAt(j)
		if predicate(r) {
			return j
		}
		j += rlen
	}
	return j
}

func (b *linebreaker[T]) runeAt(startAt int) (rune, int) {
	return b.decoder(b.s, startAt)
}

func defaultMaxWidth() int {
	if !utils.Interactive() {
		return 0
	}
	width, _, err := terminal.GetSize(0)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to identify terminal width for line wrapping; will not wrap")
		return 0
	}
	log.Warn().Msgf("max terminal width: %v", width)
	return width
}

package wordwrap

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/leaanthony/go-ansi-parser"
	"github.com/rs/zerolog/log"
	"strings"
	"unicode"
	"unicode/utf8"
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
	FixedMaxWidth              int    // FixedMaxWidth maximum width of line / length at which newline should be inserted
	DynamicMaxWidth            bool   // DynamicMaxWidth if true, detect terminal max width before every wrap
	Padding                    string // Padding string to pad wrapped lines with. Eg. "  " for 2-space indent
	EscapeNewlineStringLiteral bool   // EscapeNewlineStringLiteral For `"long strings like this"`, indicates whether inserted newlines should be prefixed with ` \`
}

func (o *Options) maxWidth() int {
	if o.DynamicMaxWidth {
		return utils.TerminalWidth()
	}
	return o.FixedMaxWidth
}

type Wrapper interface {
	WrapTo(s string, maxWidth int) string
	Wrap(s string) string
}

func New(opts ...func(*Options)) Wrapper {
	options := utils.CollateOptions(Options{
		FixedMaxWidth:              utils.TerminalWidth(),
		DynamicMaxWidth:            false,
		Padding:                    "",
		EscapeNewlineStringLiteral: false,
	}, opts...)

	return wrapper{
		options:   options,
		padOffset: printableLength(options.Padding),
	}
}

// wrapper implements the wordWrapper interface
type wrapper struct {
	options   Options
	padOffset int // number of printable characters in options.Padding
}

func (w wrapper) Wrap(s string) string {
	return w.WrapTo(s, w.options.maxWidth())
}

func (w wrapper) WrapTo(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	return newLineBreaker(s, maxWidth, w).wrap()
}

type linebreaker struct {
	padding                    string          // optional padding to prefixed wrapped lines with
	escapeNewlineStringLiteral bool            // whether newlines enclosed in quotes should be escaped
	maxWidth                   int             // maximum length of line (in printable characters)
	s                          string          // the original, unwrapped string
	buf                        strings.Builder // stores the word-wrapped string we're building as we traverse s
	i                          int             // in s: start of current word, and the first byte not yet copied to buf
	n                          int             // in s: start of the next word
	quoteState                 quoteState      // tracks whether current word is inside unescaped quotation marks
	lineOffset                 int             // in buf: number of characters written since last line break
	paddingOffset              int             // number of printable characters in Padding
}

func newLineBreaker(input string, maxWidth int, w wrapper) *linebreaker {
	return &linebreaker{
		maxWidth:                   maxWidth,
		escapeNewlineStringLiteral: w.options.EscapeNewlineStringLiteral,
		padding:                    w.options.Padding,
		s:                          input,
		i:                          0,
		n:                          0,
		lineOffset:                 0,
		paddingOffset:              w.padOffset,
	}
}

func (b *linebreaker) wrap() string {
	for b.advanceToNextWord() {
		if b.wordExceedsMaxWidth() && !b.atStartOfLine() {
			b.startNewLine()
		}
		b.copyWordToBuffer()
	}

	return b.buf.String()
}

func (b *linebreaker) currentWord() string {
	return b.s[b.i:b.n]
}

// copy current word to buffer, incrementing line offset as we go
func (b *linebreaker) copyWordToBuffer() {
	word := b.currentWord()
	b.buf.WriteString(word)

	// scan word for newline - if one existed, set lineOffset back to 0
	i := strings.LastIndex(word, "\n")
	if i != -1 {
		b.lineOffset = 0
	}
	b.lineOffset += printableLength(word[i+1:])
}

func (b *linebreaker) startNewLine() {
	if b.shouldEscapeNewlineBeforeWrapping() {
		// add ` \` to end of current line before adding a new line
		_ = b.buf.WriteByte(space)
		_ = b.buf.WriteByte(backslash)
	}

	// if the first character of the current word is whitespace, replace it
	// with the newline by incrementing i before the word is copied.
	// This way wrapped lines will not start with a leading space.
	if r, _ := utf8.DecodeRuneInString(b.s[b.i:]); unicode.IsSpace(r) {
		b.i++
	}

	// write newline and reset line offset
	_ = b.buf.WriteByte(newLine)
	b.lineOffset = 0

	// write optional padding after newline, but only if the current word contains non-whitespace
	if !b.currentWordWhitespaceOnly() {
		_, _ = b.buf.WriteString(b.padding)
		b.lineOffset += b.paddingOffset
	}
}

func (b *linebreaker) wordExceedsMaxWidth() bool {
	maxWidth := b.maxWidth
	if b.shouldEscapeNewlineBeforeWrapping() {
		// take off 2 chars to leave room for escape sequence
		maxWidth -= 2
	}
	return (b.lineOffset + b.wordLength()) > maxWidth
}

func (b *linebreaker) wordLength() int {
	return printableLength(b.currentWord())
}

func (b *linebreaker) atStartOfLine() bool {
	return b.lineOffset == 0
}

func (b *linebreaker) currentWordWhitespaceOnly() bool {
	for _, r := range b.currentWord() {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// move on to the next word, returning false if we've reached end of s
func (b *linebreaker) advanceToNextWord() bool {
	b.i = b.n
	b.n = b.findStartOfNextWord(b.n)

	b.countQuotes()

	return b.i != b.n
}

func (b *linebreaker) shouldEscapeNewlineBeforeWrapping() bool {
	if !b.escapeNewlineStringLiteral {
		return false
	}
	return b.quoteState == inside || b.quoteState == ends
}

// scan through substring and check for quote marks
func (b *linebreaker) countQuotes() {
	var count int
	var prev rune

	// iterate through runes in current word
	for j, r := range b.currentWord() {
		if r == doubleQuote {
			// make sure previous rune was not a \
			if j == 0 || prev != backslash {
				count++
			}
		}
		prev = r
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

func (b *linebreaker) findStartOfNextWord(startAt int) int {
	search := b.s[startAt:]

	var foundNonSpace bool

	for j, r := range search {
		if !foundNonSpace {
			if unicode.IsSpace(r) {
				continue
			}
			foundNonSpace = true
			continue
		}

		if !unicode.IsSpace(r) {
			continue
		}
		return startAt + j
	}

	return len(b.s)
}

func printableLength(s string) int {
	clean, err := ansi.Cleanse(s)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to strip ansi control characters from string")
	} else {
		s = clean
	}

	var ln int
	for _, r := range s {
		if unicode.IsPrint(r) {
			ln++
		}
	}

	return ln
}

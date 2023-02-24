package logging

import (
	"fmt"
	"strings"
	"unicode"
)

//
//// WordWrappingWriter a custom io.Writer that word-wraps
//// long lines to current terminal width
//type WordWrappingWriter struct {
//	// inner writer to send log messages to
//	inner zerolog.LevelWriter
//	// maxWidth length at which lines should be word-wrapped
//	maxWidth int
//	// wrapPrefix optional prefix to prepend to wrapped lines
//	wrapPrefix []byte
//}
//
//func NewWordWrappingWriter(inner zerolog.LevelWriter, maxWidth int, wrapPrefix string) zerolog.LevelWriter {
//	return &WordWrappingWriter{
//		inner:      inner,
//		maxWidth:   maxWidth,
//		wrapPrefix: []byte(wrapPrefix),
//	}
//}
//
//func (w *WordWrappingWriter) Write(p []byte) (n int, err error) {
//	panic("TODO")
//	return w.inner.Write(p)
//}
//
//func (w *WordWrappingWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
//	panic("TODO")
//	return w.inner.WriteLevel(level, p)
//}
//
//type wrapper struct {
//	input      string
//	i          int
//	p          int
//	maxWidth   int
//	lineOffset int
//	buf        strings.Builder
//}

func wrap(input string, maxWidth int, padding string, escapeStringLiteral bool) string {
	s := []rune(input)
	var buf strings.Builder
	i := 0  // i represents end of the current word
	p := -1 // p represents end of the last word. s[p] will always be the recently copied to the buffer
	inQuotes := false
	lineOffset := 0 // offset in buf since start of the last line

	// until we have copied the last character
	for p < len(s)-1 {
		i = findEndOfNextWord(s, p)
		fmt.Printf("P: %d, I: %d\n", p, i)

		wordlen := i - p

		// for long words at the start of a line,
		// copy the entire word to the buffer,
		// let the terminal wrap it, and move on.
		if lineOffset == 0 && wordlen > maxWidth {
			copyToBuf(s, &buf, p, i)
			if i != len(s)-1 {
				// if we aren't at the end of the string, then replace first space after i with a newline
				buf.WriteByte('\n')
				i += 1 // increment to skip
			}
			p = i
			continue
		}

		if containsOddNumberOfQuotes(s, p, i) {
			inQuotes = !inQuotes
		}

		if (escapeStringLiteral && inQuotes && (lineOffset+wordlen) > (maxWidth-2)) ||
			((lineOffset + wordlen) > maxWidth) {

			if escapeStringLiteral {
				buf.WriteRune(s[p+1])
				buf.WriteByte('\\')
			}

			// replace first space after p with a newline
			buf.WriteByte('\n')
			p++

			// write padding after newline
			buf.WriteString(padding)
			lineOffset = len(padding)
		}

		copyToBuf(s, &buf, p, i)
		lineOffset += i - p
		p = i
	}

	return buf.String()
}

func copyToBuf(s []rune, buf *strings.Builder, startAfter int, endAt int) {
	for k := startAfter + 1; k <= endAt; k++ {
		fmt.Printf("WRITING RUNE AT %d\n", k)
		buf.WriteRune(s[k])
	}
}

// return true if substring contains an odd number of double quote '"' characters
func containsOddNumberOfQuotes(s []rune, startAfter int, endAt int) bool {
	count := 0
	for i := startAfter + 1; i <= endAt; i++ {
		if s[i] == '"' {
			count++
		}
	}
	return count%2 == 1
}

func findEndOfNextWord(s []rune, startAfter int) int {
	// ignore leading whitespace, that counts as part of the word
	i := startAfter + 1
	for i < len(s) && unicode.IsSpace(s[i]) {
		i++
	}
	for i < len(s) && !unicode.IsSpace(s[i]) {
		i++
	}
	return i - 1
}

//
//func wordWrap(message []byte) []byte {
//	// we need to convert this bad boy to a string
//	// convert to string so we can operate on runes instead of bytes
//	// okay so - if a message has "  foo  "- we want to treat that like a single word, right?
//	// that's a difference here.
//	// INF attempt to run command "this is a very long
//	//     command that needs a wrap" unfortunately failed
//	//
//	// okay yeah splitting seems fine.
//	// do we need to preserve additional (n+1) spaces? seems like a good thing to do.
//	// so if I have "echo foo    bar" that becomes
//	// INF sadly the command "echo foo
//	//         bar" failed
//	// so one option would be to add a \
//	// to things in quotes that are split up over multiple lines.
//	// INF sadly the command "echo 'foo \
//	//     bar" failed
//	//
//	// INF the very long message "here is \
//	//     a long message" was printed
//	//
//	// INF the very long message "here is a lo
//	// ng message" was printed
//	//
//	// INF this is a long message that does
//	//     not include a quote
//	// OKAY this works since whitespace is not usually significant.
//	//
//	// We can make the prefix configurable.
//	// THELMA_LOGGING_CONSOLE_WORDWRAP_ENABLED
//	// THELMA_LOGGING_CONSOLE_WORDWRAP_PAD_WRAPPED_LINE
//	// THELMA_LOGGING_CONSOLE_WORDWRAP_BACKSLASH_STRING_LITERAL
//	//
//	// also, ideally we use the same logic in prompt.
//	//
//	// okay so we insert newlines (and maybe a backslash)
//	// at the end of a word, REPLACING ONE (1) SPACE WITH A NEWLINE.
//	//
//	// if we are in quotes, we replace one space with 3 characters: " \\\n"
//	//
//	// we have an end-of-last-word counter p (which is known to be within max len and has been copied).
//	// then we find end-of-next-word i.
//	//
//	// i = findEndOfNextWord(s, p)
//	//
//	// word = s[p+1:i]
//	// len = len(word)
//	//
//	// if word includes an unescaped quotation mark
//	//   toggle inQuotes
//	//
//	// if inQuotes && (lineOffset + len) < (maxWidth - 2)
//	//   buf += " \\\n"
//	//   buf += word
//	//   lineOffset = len
//	//   p = i
//	// else if (lineOffset + len) < maxWidth
//	//   buf += "\n"
//	// 	 buf += word
//	//	 lineOffset = len
//	//   p = i
//	// else
//	//   buf += word
//	//   lineOffset += len
//	//   p = i
//	//
//	// if inQuotes && p < (maxWidth - 2)
//	//
//	//           |
//	// 01234567890123
//	// "a long q"
//	//           |
//	// 01234567890123
//	// "a long q "
//	// "a long \
//	// q "
//	//
//	// "a long quote"
//	//
//	//
//	// message "foo \"bar \\"baz\\" quux\" blah"
//	m := string(message)
//
//	var wrapped []byte
//	var count int
//	for i := 0; i < len(message); i++ {
//		unicode.IsSpace(message[i])
//	}
//}

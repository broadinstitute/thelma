package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Wrap(t *testing.T) {
	assert.Equal(t, "", wrap("", 10, "", false))
	assert.Equal(t, "a", wrap("a", 10, "", false))
	assert.Equal(t, "abcd", wrap("abcd", 10, "", false))
	assert.Equal(t, "abcde", wrap("abcde", 5, "", false))

	assert.Equal(t, "a b c\nd e f", wrap("a b c d e f", 5, "", false))
	assert.Equal(t, "a b c\nd e f\n", wrap("a b c d e f ", 5, "", false))
	assert.Equal(t, "a b c\nd e f\n ", wrap("a b c d e f  ", 5, "", false))
	assert.Equal(t, "a b c\nd e f\ng", wrap("a b c d e f g", 5, "", false))

	assert.Equal(t, "the\nquick\nbrown\nfox\njumped\nover\nthe\nlazy\ndog", wrap("the quick brown fox jumped over the lazy dog", 5, "", false))
	assert.Equal(t, "the quick\nbrown fox\njumped\nover the\nlazy dog", wrap("the quick brown fox jumped over the lazy dog", 10, "", false))

	// longer-than-max-width-lines should be preserved on one line.
	assert.Equal(t, "abcdef", wrap("abcdef", 5, "", false))
	assert.Equal(t, "abcdefghijklmnopqrstuv", wrap("abcdefghijklmnopqrstuv", 5, "", false))
	assert.Equal(t, "          ", wrap("          ", 5, "", false))

	assert.Equal(t, "abcdef\n", wrap("abcdef ", 5, "", false))
	assert.Equal(t, "abcdef\n", wrap("abcdef\n", 5, "", false))
	assert.Equal(t, "abcdef\n ", wrap("abcdef\n ", 5, "", false))
	assert.Equal(t, "abcdef\n\n", wrap("abcdef \n", 5, "", false))
}

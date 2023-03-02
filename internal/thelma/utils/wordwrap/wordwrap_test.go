package wordwrap

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Wordwrap(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		maxWidth     int
		padding      string
		escapeQuotes bool
		expected     string
	}{
		{
			name:     "empty",
			input:    "",
			maxWidth: 5,
			expected: "",
		},
		{
			name:     "maxwidth 0 disables wrapping",
			input:    "a",
			maxWidth: 0,
			expected: "a",
		},
		{
			input:    "a",
			maxWidth: 1,
			expected: "a",
		},
		{
			input:    "a",
			maxWidth: 5,
			expected: "a",
		},
		{
			input:    "abcd",
			maxWidth: 5,
			expected: "abcd",
		},
		{
			input:    "abcde",
			maxWidth: 5,
			expected: "abcde",
		},
		{
			name:     "long words should not be wrapped",
			input:    "abcdef",
			maxWidth: 5,
			expected: "abcdef",
		},
		{
			input:    "a b c d e f",
			maxWidth: 0,
			expected: "a b c d e f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 1,
			expected: "a\nb\nc\nd\ne\nf",
		},
		{
			input:    "a b c d e f",
			maxWidth: 2,
			expected: "a\nb\nc\nd\ne\nf",
		},
		{
			input:    "a b c d e f",
			maxWidth: 3,
			expected: "a b\nc d\ne f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 4,
			expected: "a b\nc d\ne f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 5,
			expected: "a b c\nd e f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 6,
			expected: "a b c\nd e f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 7,
			expected: "a b c d\ne f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 8,
			expected: "a b c d\ne f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 9,
			expected: "a b c d e\nf",
		},
		{
			input:    "a b c d e f",
			maxWidth: 10,
			expected: "a b c d e\nf",
		},
		{
			input:    "a b c d e f",
			maxWidth: 11,
			expected: "a b c d e f",
		},
		{
			input:    "a b c d e f",
			maxWidth: 12,
			expected: "a b c d e f",
		},
		{
			name:     "quick brown fox, max=0",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 0,
			expected: "the quick brown fox jumped over the lazy dog",
		},
		{
			name:     "quick brown, max=1",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 1,
			expected: "the\nquick\nbrown\nfox\njumped\nover\nthe\nlazy\ndog",
		},
		{
			name:     "quick brown, max=5",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 5,
			expected: "the\nquick\nbrown\nfox\njumped\nover\nthe\nlazy\ndog",
		},
		{
			name:     "quick brown fox, max=10",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 10,
			expected: "the quick\nbrown fox\njumped\nover the\nlazy dog",
		},
		{
			name:     "quick brown fox, max=20",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 20,
			expected: "the quick brown fox\njumped over the lazy\ndog",
		},
		{
			name:     "quick brown fox, max=40",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 40,
			expected: "the quick brown fox jumped over the lazy\ndog",
		},
		{
			name:     "quick brown fox, max=80",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 80,
			expected: "the quick brown fox jumped over the lazy dog",
		},
		{
			name:     "quick brown fox, max=0, padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 0,
			padding:  " > ",
			expected: "the quick brown fox jumped over the lazy dog",
		},
		{
			name:     "quick brown, max=1, padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 1,
			padding:  " > ",
			expected: "the\n > quick\n > brown\n > fox\n > jumped\n > over\n > the\n > lazy\n > dog",
		},
		{
			name:     "quick brown, max=5, padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 5,
			padding:  " > ",
			expected: "the\n > quick\n > brown\n > fox\n > jumped\n > over\n > the\n > lazy\n > dog",
		},
		{
			name:     "quick brown fox, max=10, padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 10,
			padding:  " > ",
			expected: "the quick\n > brown\n > fox\n > jumped\n > over\n > the\n > lazy\n > dog",
		},
		{
			name:     "quick brown fox, max=20, padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 20,
			padding:  " > ",
			expected: "the quick brown fox\n > jumped over the\n > lazy dog",
		},
		{
			name:     "quick brown fox, max=40, padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 40,
			padding:  " > ",
			expected: "the quick brown fox jumped over the lazy\n > dog",
		},
		{
			name:     "quick brown fox, max=80, padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 80,
			padding:  " > ",
			expected: "the quick brown fox jumped over the lazy dog",
		},
		{
			name:     "quick brown fox, max=20, long padding",
			input:    "the quick brown fox jumped over the lazy dog",
			maxWidth: 20,
			padding:  "          > ",
			expected: "the quick brown fox\n          > jumped\n          > over the\n          > lazy dog",
		},
		{
			name: "long paragraphs should be wrapped",
			input: `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed bibendum tempor libero ac egestas. Nam convallis odio quis ex vulputate, non vestibulum erat auctor. Mauris sollicitudin volutpat elit, sit amet consequat mi accumsan vel. Praesent quis sollicitudin arcu. Aenean vitae iaculis purus. Donec sollicitudin dui a luctus porttitor. In sollicitudin euismod libero quis varius. Nunc tincidunt eget dolor id rutrum. Phasellus consequat massa vitae justo hendrerit feugiat. Proin at tristique erat. Vivamus tincidunt ipsum vitae ipsum porttitor pharetra. Morbi sit amet ligula pellentesque libero tempus interdum vitae ac metus. Maecenas dignissim ipsum nulla, vitae fringilla massa vehicula eu. Vivamus eu diam nibh.

Nullam accumsan vestibulum ipsum commodo scelerisque. Morbi viverra arcu felis, et ornare lectus sollicitudin ac. Sed id iaculis tortor. Proin viverra consectetur posuere. Quisque elit dui, consectetur et est nec, vehicula varius nulla. Vivamus rhoncus semper justo, eget tincidunt lectus. Maecenas malesuada ultrices erat at pharetra. Nulla facilisi. Sed suscipit, turpis vel ultrices vestibulum, diam purus tincidunt sapien, ut pellentesque ipsum metus nec nisi. Cras aliquam, nibh id placerat posuere, sapien elit iaculis velit, vel dictum nulla augue et turpis. Donec at aliquet velit. Ut dignissim, orci sit amet commodo volutpat, augue enim cursus nunc, non venenatis est mauris ultricies dolor. Phasellus vehicula enim ut augue ultrices egestas. Cras efficitur libero at justo pretium, sollicitudin varius ante mattis. Vivamus in sodales lorem. Mauris efficitur sit amet nisl in euismod.

Phasellus nec est posuere, condimentum lacus non, porttitor justo. Nullam dapibus vitae dui ac vulputate. Interdum et malesuada fames ac ante ipsum primis in faucibus. Nullam aliquet dictum tellus vitae ullamcorper. Phasellus nec vestibulum dolor, id pretium mauris. Sed vestibulum sodales sapien sed porta. Donec mauris neque, dapibus sed convallis vel, efficitur quis magna. Suspendisse orci nibh, sagittis vitae risus sed, bibendum blandit diam.

Cras eu bibendum nunc. Fusce tortor odio, pulvinar quis iaculis sed, posuere sit amet nibh. Nam mollis sapien et diam facilisis feugiat. Mauris et iaculis mauris, non tincidunt nisl. Sed in diam sagittis, convallis purus et, eleifend nisl. Curabitur quis ultricies tellus, at porttitor leo. Quisque ut nunc condimentum, euismod leo et, rhoncus mi.

Morbi pharetra nisi eleifend nunc placerat dictum ut maximus est. Nulla facilisi. Aenean porttitor vel leo eu egestas. Nam at nunc semper, vulputate dolor et, tempor ex. Fusce ut tellus commodo, tincidunt mauris sit amet, laoreet odio. Sed fermentum aliquet metus ut iaculis. Integer sit amet felis orci. Vestibulum nec elit vitae tellus volutpat cursus finibus et ligula. Nulla ullamcorper quam a augue ultrices dignissim. Nam ac porttitor eros. Pellentesque quis neque nec nibh mattis vehicula in id ipsum. Vestibulum in tempus lorem, in congue mi. Vestibulum posuere, erat eu congue ullamcorper, risus augue condimentum risus, vel vulputate mauris neque sit amet tellus. Nunc et turpis tristique libero luctus luctus a sit amet urna. Sed ut nulla a urna vestibulum pellentesque eget vel massa. Sed tincidunt placerat augue id cursus.
`,
			maxWidth: 80,
			expected: `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed bibendum tempor
libero ac egestas. Nam convallis odio quis ex vulputate, non vestibulum erat
auctor. Mauris sollicitudin volutpat elit, sit amet consequat mi accumsan vel.
Praesent quis sollicitudin arcu. Aenean vitae iaculis purus. Donec sollicitudin
dui a luctus porttitor. In sollicitudin euismod libero quis varius. Nunc
tincidunt eget dolor id rutrum. Phasellus consequat massa vitae justo hendrerit
feugiat. Proin at tristique erat. Vivamus tincidunt ipsum vitae ipsum porttitor
pharetra. Morbi sit amet ligula pellentesque libero tempus interdum vitae ac
metus. Maecenas dignissim ipsum nulla, vitae fringilla massa vehicula eu.
Vivamus eu diam nibh.

Nullam accumsan vestibulum ipsum commodo scelerisque. Morbi viverra arcu felis,
et ornare lectus sollicitudin ac. Sed id iaculis tortor. Proin viverra
consectetur posuere. Quisque elit dui, consectetur et est nec, vehicula varius
nulla. Vivamus rhoncus semper justo, eget tincidunt lectus. Maecenas malesuada
ultrices erat at pharetra. Nulla facilisi. Sed suscipit, turpis vel ultrices
vestibulum, diam purus tincidunt sapien, ut pellentesque ipsum metus nec nisi.
Cras aliquam, nibh id placerat posuere, sapien elit iaculis velit, vel dictum
nulla augue et turpis. Donec at aliquet velit. Ut dignissim, orci sit amet
commodo volutpat, augue enim cursus nunc, non venenatis est mauris ultricies
dolor. Phasellus vehicula enim ut augue ultrices egestas. Cras efficitur libero
at justo pretium, sollicitudin varius ante mattis. Vivamus in sodales lorem.
Mauris efficitur sit amet nisl in euismod.

Phasellus nec est posuere, condimentum lacus non, porttitor justo. Nullam
dapibus vitae dui ac vulputate. Interdum et malesuada fames ac ante ipsum primis
in faucibus. Nullam aliquet dictum tellus vitae ullamcorper. Phasellus nec
vestibulum dolor, id pretium mauris. Sed vestibulum sodales sapien sed porta.
Donec mauris neque, dapibus sed convallis vel, efficitur quis magna. Suspendisse
orci nibh, sagittis vitae risus sed, bibendum blandit diam.

Cras eu bibendum nunc. Fusce tortor odio, pulvinar quis iaculis sed, posuere sit
amet nibh. Nam mollis sapien et diam facilisis feugiat. Mauris et iaculis
mauris, non tincidunt nisl. Sed in diam sagittis, convallis purus et, eleifend
nisl. Curabitur quis ultricies tellus, at porttitor leo. Quisque ut nunc
condimentum, euismod leo et, rhoncus mi.

Morbi pharetra nisi eleifend nunc placerat dictum ut maximus est. Nulla
facilisi. Aenean porttitor vel leo eu egestas. Nam at nunc semper, vulputate
dolor et, tempor ex. Fusce ut tellus commodo, tincidunt mauris sit amet, laoreet
odio. Sed fermentum aliquet metus ut iaculis. Integer sit amet felis orci.
Vestibulum nec elit vitae tellus volutpat cursus finibus et ligula. Nulla
ullamcorper quam a augue ultrices dignissim. Nam ac porttitor eros. Pellentesque
quis neque nec nibh mattis vehicula in id ipsum. Vestibulum in tempus lorem, in
congue mi. Vestibulum posuere, erat eu congue ullamcorper, risus augue
condimentum risus, vel vulputate mauris neque sit amet tellus. Nunc et turpis
tristique libero luctus luctus a sit amet urna. Sed ut nulla a urna vestibulum
pellentesque eget vel massa. Sed tincidunt placerat augue id cursus.
`,
		},
		{
			name:         "string escaping -- counter case",
			maxWidth:     10,
			escapeQuotes: false,
			input:        `"a very long quote with spaces"`,
			expected: `"a very
long quote
with
spaces"`,
		},
		{
			name:         "string escaping -- simple",
			maxWidth:     10,
			escapeQuotes: true,
			input:        `"a very long quote with spaces"`,
			expected: `"a very \
long \
quote \
with \
spaces"`,
		},
		{
			name:         "string escaping with quoted sections",
			input:        `this "text is in quotes" but then the rest of this text is not "but hey we are in quotes again" and only the stuff in quotes should be escaped.`,
			escapeQuotes: true,
			maxWidth:     20,
			expected: `this "text is in \
quotes" but then the
rest of this text is
not "but hey we \
are in quotes \
again" and only the
stuff in quotes
should be escaped.`,
		},
		{
			name:         "string escaping with nested escaped quotes",
			input:        `okay so this part of the text is not quoted "but this part is and also \"it has some escaped interior quotes \\\"and look, yet another layer of nesting!\\\" but the entire thing within outer quotes should be escaped,\" okay?" and then this part should not be`,
			escapeQuotes: true,
			maxWidth:     20,
			expected: `okay so this part of
the text is not
quoted "but this \
part is and also \
\"it has some \
escaped interior \
quotes \\\"and \
look, yet another \
layer of \
nesting!\\\" but \
the entire thing \
within outer \
quotes should be \
escaped,\" okay?"
and then this part
should not be`,
		},
		{
			name:         "string escaping long command with quotes",
			maxWidth:     30,
			escapeQuotes: true,
			input:        `command failed: "/usr/bin/env VAR1=some-long-string VAR2=another-long-string VAR3=yet-another-long-string /a/very/very/very/very/long/path/to/my/program --foo=a-very-long-argument --inserting-a-newline-here-will-break-command-copy-pasting --which-is-annoying"`,
			expected: `command failed: "/usr/bin/env \
VAR1=some-long-string \
VAR2=another-long-string \
VAR3=yet-another-long-string \
/a/very/very/very/very/long/path/to/my/program \
--foo=a-very-long-argument \
--inserting-a-newline-here-will-break-command-copy-pasting \
--which-is-annoying"`,
		},
		{
			name:         "string escaping long command with quotes and padding",
			maxWidth:     30,
			escapeQuotes: true,
			padding:      "    ",
			input:        `command failed: "/usr/bin/env VAR1=some-long-string VAR2=another-long-string VAR3=yet-another-long-string /a/very/very/very/very/long/path/to/my/program --foo=a-very-long-argument --inserting-a-newline-here-will-break-command-copy-pasting --which-is-annoying"`,
			expected: `command failed: "/usr/bin/env \
    VAR1=some-long-string \
    VAR2=another-long-string \
    VAR3=yet-another-long-string \
    /a/very/very/very/very/long/path/to/my/program \
    --foo=a-very-long-argument \
    --inserting-a-newline-here-will-break-command-copy-pasting \
    --which-is-annoying"`,
		},
		{
			name:     "ansii escape sequences should not count towards line length when wrapping",
			maxWidth: 9,
			input:    "\u001b[31mRED\u001b[0m \u001b[32mGREEN\u001b[0m \u001b[33mYELLOW\u001b[0m",
			expected: "\u001b[31mRED\u001b[0m \u001b[32mGREEN\u001b[0m\n\u001b[33mYELLOW\u001b[0m",
		},
		{
			name:     "long word at end of line should not add padding after line",
			maxWidth: 15,
			input:    "long word occursatendofstring\n",
			padding:  "      ",
			expected: "long word\n      occursatendofstring\n",
		},
		{
			name:     "should correctly wrap emoji",
			maxWidth: 10,
			input:    "🦄🦄🦄 a 👹 b 🎏c🎏d🎏 e 🦜f🦜g🦜h🦜i🦜j🦜k ⚠️ ☣️ l m 💯 💯 🚽 🧻",
			padding:  "🍜🍜🍜🍜",
			expected: `🦄🦄🦄 a 👹 b
🍜🍜🍜🍜🎏c🎏d🎏
🍜🍜🍜🍜e
🍜🍜🍜🍜🦜f🦜g🦜h🦜i🦜j🦜k
🍜🍜🍜🍜⚠️ ☣️
🍜🍜🍜🍜l m 💯
🍜🍜🍜🍜💯 🚽 🧻`,
		},
	}

	for _, tc := range testCases {
		name := tc.name
		if name == "" {
			name = fmt.Sprintf("%q (max=%d padding=%q esc=%t)", tc.input, tc.maxWidth, tc.padding, tc.escapeQuotes)
		}
		t.Run(name, func(t *testing.T) {
			opts := func(options *Options) {
				options.FixedMaxWidth = tc.maxWidth
				options.Padding = tc.padding
				options.EscapeNewlineStringLiteral = tc.escapeQuotes
			}
			w := New(opts)

			assert.Equal(t, tc.expected, w.Wrap(tc.input))
		})
	}
}

package format

import (
	"bytes"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog/log"
	"io"
)

// Name of the Chroma YAML lexer
const chromaYamlLexer = "YAML"

// Chroma styles for light and dark terminal themes
// see https://swapoff.org/chroma/playground/ for full list
const chromaStyleLight = "friendly"
const chromaStyleDark = "dracula"

func formatPrettyYaml(data interface{}, w io.Writer) error {
	return formatPrettyYamlWithOptions(data, w, chooseStyle(), chooseFormatter())
}

func formatPrettyYamlWithOptions(data interface{}, w io.Writer, styleName string, formatterName string) error {
	// first, convert to YAML (chroma won't do that for us)
	var b bytes.Buffer
	if err := formatYaml(data, &b); err != nil {
		return err
	}
	content := b.Bytes()

	// following chroma docs
	// https://github.com/alecthomas/chroma#formatting-the-output
	lexer := lexers.Get(chromaYamlLexer)
	if lexer == nil {
		log.Warn().Msgf("Couldn't load chroma lexer %q, falling back to plain yaml format", chromaYamlLexer)
		return formatYaml(data, w)
	}

	style := styles.Get(styleName)
	if style == nil {
		log.Warn().Msgf("Couldn't load chroma style %q, falling back to plain yaml format", styleName)
		return formatYaml(data, w)
	}

	formatter := formatters.Get(formatterName)
	if formatter == nil {
		log.Warn().Msgf("Couldn't load chroma formatter %q, falling back to plain YAML format", formatterName)
		return formatYaml(data, w)
	}

	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, string(content))
	if err != nil {
		return err
	}

	return formatter.Format(w, style, iterator)
}

func chooseFormatter() string {
	// available Chroma formatters described here:
	//   https://pkg.go.dev/github.com/alecthomas/chroma/formatters
	switch termenv.ColorProfile() {
	case termenv.ANSI:
		return "terminal16"
	case termenv.ANSI256:
		return "terminal256"
	case termenv.TrueColor:
		return "terminal16m"
	default:
		return "" // fall back to plain YAML
	}
}

func chooseStyle() string {
	if termenv.HasDarkBackground() {
		return chromaStyleDark
	} else {
		return chromaStyleLight
	}
}

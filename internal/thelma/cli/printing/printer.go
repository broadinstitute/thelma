package printing

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/cli/printing/format"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"io"
	"os"
)

const stdoutFlagVal = "STDOUT"

var flagNames = struct {
	outputFile   string
	outputFormat string
}{
	outputFile:   "output-file",
	outputFormat: "output-format",
}

type options struct {
	outputFile   string
	outputFormat string
}

// Printer is a tool for printing formatted output from thelma cli commands.
type Printer interface {
	// AddFlags adds --output-format and --output-file flags to a cobra FlagSet
	AddFlags(flags *pflag.FlagSet)
	// VerifyFlags verifies flags have correct values
	VerifyFlags() error
	// PrintOutput prints output in the user-supplied format
	PrintOutput(output interface{}, stdout io.Writer) error
}

func NewPrinter() Printer {
	return &printer{}
}

// Implements printer interface
type printer struct {
	options options
}

func (p *printer) AddFlags(flags *pflag.FlagSet) {
	defaultFormat := format.PrettyYaml
	if !utils.Interactive() {
		// use plain YAML if this is not an interactive terminal
		defaultFormat = format.Yaml
	}
	flags.StringVar(&p.options.outputFormat, flagNames.outputFormat, defaultFormat.String(), fmt.Sprintf("One of: %s", utils.QuoteJoin(format.SupportedFormats())))
	flags.StringVar(&p.options.outputFile, flagNames.outputFile, stdoutFlagVal, "Optionally write output to file instead of stdout")
}

func (p *printer) VerifyFlags() error {
	if !format.IsSupported(p.options.outputFormat) {
		return fmt.Errorf("--%s must be one of %s; got %q", flagNames.outputFormat, utils.QuoteJoin(format.SupportedFormats()), p.options.outputFormat)
	}
	return nil
}

func (p *printer) PrintOutput(output interface{}, stdout io.Writer) error {
	var ft format.Format
	if err := (&ft).FromString(p.options.outputFormat); err != nil {
		return err
	}

	if p.options.outputFile == stdoutFlagVal {
		// write to stdout
		return ft.Format(output, stdout)
	} else {
		// write to file
		return printToFile(output, ft, p.options.outputFile)
	}
}

func printToFile(output interface{}, ft format.Format, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err = ft.Format(output, f); err != nil {
		if err2 := f.Close(); err2 != nil {
			log.Err(err2)
		}
		return err
	}

	return f.Close()
}

package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/spf13/cobra"
	"io"
)

// Option configuration option for a ThelmaCLI
type Option func(*Options)

// Options configuration options for a ThelmaCLI
type Options struct {
	commands         map[string]ThelmaCommand
	thelmaConfigHook func(builder.ThelmaBuilder)
	cobraConfigHooks []func(*cobra.Command)
	skipRun          bool
}

// DefaultOptions default options
func DefaultOptions() *Options {
	return &Options{
		commands: make(map[string]ThelmaCommand),
	}
}

// AddCommand add a subcommand to the ThelmaCLI
func (opts *Options) AddCommand(name string, cmd ThelmaCommand) {
	if err := validateCommandName(name); err != nil {
		panic(err)
	}
	opts.commands[name] = cmd
}

// ConfigureThelma add a configration hook for Thelma
func (opts *Options) ConfigureThelma(hook func(builder.ThelmaBuilder)) {
	opts.thelmaConfigHook = hook
}

// SetOut write output to the given writer instead of os.Stdout
func (opts *Options) SetOut(writer io.Writer) {
	opts.cobraConfigHooks = append(opts.cobraConfigHooks, func(command *cobra.Command) {
		command.SetOut(writer)
	})
}

// SetArgs set CLI args to the given list instead of os.Args
func (opts *Options) SetArgs(args []string) {
	opts.cobraConfigHooks = append(opts.cobraConfigHooks, func(command *cobra.Command) {
		command.SetArgs(args)
	})
}

// SkipRun if true, skip run phase and only execute pre/post run hooks
func (opts *Options) SkipRun(skipRun bool) {
	opts.skipRun = skipRun
}

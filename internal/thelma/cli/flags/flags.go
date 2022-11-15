// Package flags contains common utilities for interacting with CLI flags
package flags

import "github.com/spf13/pflag"

type Options struct {
	Prefix      string
	NoShortHand bool
	Hidden      bool
}

type Option func(options *Options)

// AsOptions converts list of Option functions into an Options
func AsOptions(opts []Option) Options {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// Apply rewrites the flags in the FlagSet in accordance with the given options so that:
// * flag name is prefixed with the given Prefix
// * flag name's shorthand flag is removed if NoShortHand is true
// * flag is hidden if Hidden is true
func (f Options) Apply(flagSet *pflag.FlagSet) {
	flagSet.VisitAll(func(flag *pflag.Flag) {
		if f.Prefix != "" {
			flag.Name = f.Prefix + "-" + flag.Name
		}
		if f.NoShortHand {
			flag.Shorthand = ""
		}
		flag.Hidden = f.Hidden
	})
}

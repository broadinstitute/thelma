// Package flags contains common utilities for interacting with CLI flags
package flags

import (
	"github.com/spf13/pflag"
)

type Options struct {
	// Prefix prefix all flag names with a string
	Prefix string
	// NoShortHand do not add shorthand flags
	NoShortHand bool
	// Hidden mark flags as hidden (do not show in --help output)
	Hidden bool
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
func (f Options) Apply(original *pflag.FlagSet, addFn func(set *pflag.FlagSet)) {
	empty := new(pflag.FlagSet)
	addFn(empty)
	empty.VisitAll(func(flag *pflag.Flag) {
		flag.Name = f.NormalizedFlagName(flag.Name)
		if f.NoShortHand {
			flag.Shorthand = ""
		}
		flag.Hidden = f.Hidden
	})
	original.AddFlagSet(empty)
}

// NormalizedFlagName given a flag name, apply prefix if one was configured
func (f Options) NormalizedFlagName(baseName string) string {
	if f.Prefix == "" {
		return baseName
	}
	return f.Prefix + "-" + baseName
}

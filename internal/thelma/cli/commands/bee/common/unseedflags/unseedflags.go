package unseedflags

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/bee/seed"
	"github.com/spf13/cobra"
)

type Options struct {
	Prefix      string
	NoShortHand bool
	Hidden      bool
}

type Option func(options *Options)

func NewUnseedFlags(opts ...Option) UnseedFlags {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	return &unseedFlags{
		myOptions: options,
	}
}

type UnseedFlags interface {
	AddFlags(cobraCommand *cobra.Command)
	GetOptions(cobraCommand *cobra.Command) (seed.UnseedOptions, error)
}

var flagNames = struct {
	force                   string
	step1UnregisterAllUsers string
	noSteps                 string
}{
	force:                   "force",
	step1UnregisterAllUsers: "step-1-unregister-all-users",
	noSteps:                 "no-steps",
}

type unseedFlags struct {
	myOptions     Options
	unseedOptions seed.UnseedOptions
}

func (s *unseedFlags) AddFlags(cobraCommand *cobra.Command) {
	cobraCommand.Flags().BoolVar(&s.unseedOptions.Force, s.withPrefix(flagNames.force), false, "attempt to ignore errors during unseeding")
	s.maybeHide(cobraCommand, flagNames.force)

	cobraCommand.Flags().BoolVar(&s.unseedOptions.Step1UnregisterAllUsers, s.withPrefix(flagNames.step1UnregisterAllUsers), true, "unregister all user accounts with Sam")
	s.maybeHide(cobraCommand, flagNames.step1UnregisterAllUsers)

	cobraCommand.Flags().BoolVar(&s.unseedOptions.NoSteps, s.withPrefix(flagNames.noSteps), false, "convenience flag to skip all unspecified steps, which would otherwise run by default")
	s.maybeHide(cobraCommand, flagNames.noSteps)
}

func (s *unseedFlags) GetOptions(cobraCommand *cobra.Command) (seed.UnseedOptions, error) {
	flags := cobraCommand.Flags()

	// if --no-steps was supplied, disable any steps that were not explicitly enabled
	if s.unseedOptions.NoSteps {
		if !flags.Changed(s.withPrefix(flagNames.step1UnregisterAllUsers)) {
			s.unseedOptions.Step1UnregisterAllUsers = false
		}
	}

	return s.unseedOptions, nil
}

func (s *unseedFlags) addShorthand(cobraCommand *cobra.Command, unprefixedFlagName string, shortHand string) {
	if !s.myOptions.NoShortHand {
		cobraCommand.Flags().Lookup(s.withPrefix(unprefixedFlagName)).Shorthand = shortHand
	}
}

func (s *unseedFlags) maybeHide(cobraCommand *cobra.Command, unprefixedFlagName string) {
	if s.myOptions.Hidden {
		flag := cobraCommand.Flags().Lookup(s.withPrefix(unprefixedFlagName))
		flag.Hidden = true
	}
}

func (s *unseedFlags) withPrefix(flagName string) string {
	return fmt.Sprintf("%s%s", s.myOptions.Prefix, flagName)
}

package seedflags

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

func NewSeedFlags(opts ...Option) SeedFlags {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	return &seedFlags{
		myOptions: options,
	}
}

type SeedFlags interface {
	AddFlags(cobraCommand *cobra.Command)
	GetOptions(cobraCommand *cobra.Command) (seed.SeedOptions, error)
}

var flagNames = struct {
	force                    string
	step1CreateElasticsearch string
	step2RegisterSaProfiles  string
	step3AddSaSamPermissions string
	step4RegisterTestUsers   string
	step5CreateAgora         string
	step6ExtraUser           string
	noSteps                  string
	registerSelfShortcut     string
}{
	force:                    "force",
	step1CreateElasticsearch: "step-1-create-elasticsearch",
	step2RegisterSaProfiles:  "step-2-register-sa-profiles",
	step3AddSaSamPermissions: "step-3-add-sa-sam-permissions",
	step4RegisterTestUsers:   "step-4-register-test-users",
	step5CreateAgora:         "step-5-create-agora",
	step6ExtraUser:           "step-6-extra-user",
	noSteps:                  "no-steps",
	registerSelfShortcut:     "me",
}

type seedFlags struct {
	myOptions   Options
	seedOptions seed.SeedOptions
}

func (s *seedFlags) AddFlags(cobraCommand *cobra.Command) {
	cobraCommand.Flags().BoolVar(&s.seedOptions.Force, s.withPrefix(flagNames.force), false, "attempt to ignore errors during seeding")
	s.maybeHide(cobraCommand, flagNames.force)

	cobraCommand.Flags().BoolVar(&s.seedOptions.Step1CreateElasticsearch, s.withPrefix(flagNames.step1CreateElasticsearch), true, "create healthy Ontology index with Elasticsearch")
	s.maybeHide(cobraCommand, flagNames.step1CreateElasticsearch)

	cobraCommand.Flags().BoolVar(&s.seedOptions.Step2RegisterSaProfiles, s.withPrefix(flagNames.step2RegisterSaProfiles), true, "register service account profiles with Orch")
	s.maybeHide(cobraCommand, flagNames.step2RegisterSaProfiles)

	cobraCommand.Flags().BoolVar(&s.seedOptions.Step3AddSaSamPermissions, s.withPrefix(flagNames.step3AddSaSamPermissions), true, "add permissions for app service accounts in Sam")
	s.maybeHide(cobraCommand, flagNames.step3AddSaSamPermissions)

	cobraCommand.Flags().BoolVar(&s.seedOptions.Step4RegisterTestUsers, s.withPrefix(flagNames.step4RegisterTestUsers), true, "register test user accounts with Orch and accept TOS with Sam")
	s.maybeHide(cobraCommand, flagNames.step4RegisterTestUsers)

	cobraCommand.Flags().BoolVar(&s.seedOptions.Step5CreateAgora, s.withPrefix(flagNames.step5CreateAgora), true, "create Agora's methods repository with Orch")
	s.maybeHide(cobraCommand, flagNames.step5CreateAgora)

	cobraCommand.Flags().StringSliceVar(&s.seedOptions.Step6ExtraUser, s.withPrefix(flagNames.step6ExtraUser), []string{}, "optionally register extra users for log-in (skipped by default; can specify multiple times; provide `email` address, \"set-adc\", or \"use-adc\")")
	s.maybeHide(cobraCommand, flagNames.step6ExtraUser)

	cobraCommand.Flags().BoolVar(&s.seedOptions.NoSteps, s.withPrefix(flagNames.noSteps), false, "convenience flag to skip all unspecified steps, which would otherwise run by default")
	s.maybeHide(cobraCommand, flagNames.noSteps)

	cobraCommand.Flags().BoolVar(&s.seedOptions.RegisterSelfShortcut, s.withPrefix(flagNames.registerSelfShortcut), false, "shorthand for --step-6-extra-user use-adc")
	s.maybeHide(cobraCommand, flagNames.registerSelfShortcut)

	s.addShorthand(cobraCommand, flagNames.force, "f")
	s.addShorthand(cobraCommand, flagNames.step6ExtraUser, "u")
}

func (s *seedFlags) GetOptions(cobraCommand *cobra.Command) (seed.SeedOptions, error) {
	flags := cobraCommand.Flags()

	// if --no-steps was supplied, disable any steps that were not explicitly enabled
	if s.seedOptions.NoSteps {
		if !flags.Changed(s.withPrefix(flagNames.step1CreateElasticsearch)) {
			s.seedOptions.Step1CreateElasticsearch = false
		}
		if !flags.Changed(s.withPrefix(flagNames.step2RegisterSaProfiles)) {
			s.seedOptions.Step2RegisterSaProfiles = false
		}
		if !flags.Changed(s.withPrefix(flagNames.step3AddSaSamPermissions)) {
			s.seedOptions.Step3AddSaSamPermissions = false
		}
		if !flags.Changed(s.withPrefix(flagNames.step4RegisterTestUsers)) {
			s.seedOptions.Step4RegisterTestUsers = false
		}
		if !flags.Changed(s.withPrefix(flagNames.step5CreateAgora)) {
			s.seedOptions.Step5CreateAgora = false
		}
		// No need to handle step6ExtraUser; it is empty and does nothing by default
	}

	// handle --me
	if s.seedOptions.RegisterSelfShortcut {
		s.seedOptions.Step6ExtraUser = append(s.seedOptions.Step6ExtraUser, "use-adc")
	}

	return s.seedOptions, nil
}

func (s *seedFlags) addShorthand(cobraCommand *cobra.Command, unprefixedFlagName string, shortHand string) {
	if !s.myOptions.NoShortHand {
		cobraCommand.Flags().Lookup(s.withPrefix(unprefixedFlagName)).Shorthand = shortHand
	}
}

func (s *seedFlags) maybeHide(cobraCommand *cobra.Command, unprefixedFlagName string) {
	if s.myOptions.Hidden {
		flag := cobraCommand.Flags().Lookup(s.withPrefix(unprefixedFlagName))
		flag.Hidden = true
	}
}

func (s *seedFlags) withPrefix(flagName string) string {
	return fmt.Sprintf("%s%s", s.myOptions.Prefix, flagName)
}

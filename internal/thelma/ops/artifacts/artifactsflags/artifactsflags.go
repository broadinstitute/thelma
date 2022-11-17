package artifactsflags

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/cli/flags"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewArtifactsFlags(opts ...flags.Option) ArtifactsFlags {
	return &artifactsFlags{
		flagOptions: flags.AsOptions(opts),
		options:     artifacts.Options{},
	}
}

type ArtifactsFlags interface {
	AddFlags(cobraCommand *cobra.Command)
	GetOptions() (artifacts.Options, error)
}

var flagNames = struct {
	dir    string
	upload string
}{
	dir:    "dir",
	upload: "upload",
}

type artifactsFlags struct {
	flagOptions flags.Options
	options     artifacts.Options
}

func (s *artifactsFlags) AddFlags(cobraCommand *cobra.Command) {
	s.flagOptions.Apply(cobraCommand.Flags(), func(flags *pflag.FlagSet) {
		flags.StringVarP(&s.options.Dir, flagNames.dir, "d", "", "Path to local directory where artifacts should be exported")
		flags.BoolVarP(&s.options.Upload, flagNames.upload, "u", false, "If true, upload artifacts to cluster artifact bucket")
	})
}

func (s *artifactsFlags) GetOptions() (artifacts.Options, error) {
	if s.options.Dir == "" && !s.options.Upload {
		return s.options, fmt.Errorf("either --%s or --%s must be specified",
			s.flagOptions.NormalizedFlagName(flagNames.upload),
			s.flagOptions.NormalizedFlagName(flagNames.dir),
		)
	}
	return s.options, nil
}

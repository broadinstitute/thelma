package artifactsflags

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/cli/flags"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	"github.com/spf13/cobra"
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
	cobraCommand.Flags().StringVarP(&s.options.Dir, flagNames.dir, "d", "", "Path to local directory where artifacts should be exported")
	cobraCommand.Flags().BoolVarP(&s.options.Upload, flagNames.upload, "u", false, "If true, upload artifacts to cluster artifact bucket")
	s.flagOptions.Apply(cobraCommand.Flags())
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

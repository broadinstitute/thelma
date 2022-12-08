package kubeconform

import (
	"os"

	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const prog = "kubeconform"

type Kubeconform interface {
	ValidateDir(path string) error
}

type kubeconform struct {
	shell.Runner
}

func New(runner shell.Runner) *kubeconform {
	return &kubeconform{runner}
}

func (k *kubeconform) ValidateDir(path string) error {
	log.Info().Msgf("Validating rendered manifests in %s", path)
	return k.Run(shell.Command{
		Prog: prog,
		Args: []string{
			"-summary",
			"-ignore-missing-schemas",
			"-strict",
			"-output",
			"json",
			path,
		},
	}, func(opts *shell.RunOptions) {
		opts.LogLevel = zerolog.DebugLevel
		opts.Stdout = os.Stdout
		opts.Stderr = os.Stderr
	})
}

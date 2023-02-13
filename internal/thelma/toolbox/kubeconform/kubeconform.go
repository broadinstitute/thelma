package kubeconform

import (
	"os"

	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const prog = "kubeconform"

type kubeconform struct {
	shell.Runner
}

func New(runner shell.Runner) *kubeconform {
	return &kubeconform{runner}
}

// Validate Invoke kubeconform to perfom recursive manifest validation on all k8s yaml files under path
func (k *kubeconform) ValidateDir(path string) error {
	log.Info().Msgf("Validating rendered manifests in %s", path)
	return k.Run(shell.Command{
		Prog: prog,
		Args: []string{
			"-summary",
			"-ignore-missing-schemas",
			// "-strict",
			path,
		},
	}, func(opts *shell.RunOptions) {
		opts.LogLevel = zerolog.DebugLevel
		opts.Stdout = os.Stdout
		opts.Stderr = os.Stderr
	})
}

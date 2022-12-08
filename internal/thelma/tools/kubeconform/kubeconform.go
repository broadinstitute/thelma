package kubeconform

import "github.com/broadinstitute/thelma/internal/thelma/utils/shell"

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
	return k.Run(shell.Command{
		Prog: prog,
		Args: []string{
			"-summary",
			"-n 16", // number of parallel workers to use for validation
			"path",
		},
	})
}

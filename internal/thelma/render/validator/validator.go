package validator

import (
	"fmt"

	"github.com/broadinstitute/thelma/internal/thelma/tools/kubeconform"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

type Mode int

const (
	// Skip will skip performing any post render validation on k8s manifests
	Skip Mode = iota
	// Warn will print validation output but will not cause the render command to exit with error if problems are detected
	Warn
	// Fail will cause the render command to exit with an error code if validation errors are detected
	Fail
)

func FromString(value string) (Mode, error) {
	switch value {
	case "skip":
		return Skip, nil
	case "warn":
		return Warn, nil
	case "fail":
		return Fail, nil
	default:
		return Skip, fmt.Errorf("unknown validation mode: %q", value)
	}
}

func (m Mode) String() string {
	switch m {
	case Skip:
		return "skip"
	case Warn:
		return "warn"
	case Fail:
		return "fail"
	default:
		return "unknown"
	}
}

type Validator interface {
	dirValidator
	GetMode() Mode
}

type dirValidator interface {
	ValidateDir(path string) error
}

type validator struct {
	dirValidator
	Mode Mode
}

func (v validator) GetMode() Mode {
	return v.Mode
}

func New(mode Mode) validator {
	return validator{kubeconform.New(shell.NewRunner()), mode}
}

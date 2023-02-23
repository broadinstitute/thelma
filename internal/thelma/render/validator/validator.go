package validator

import (
	"fmt"

	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubeconform"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
)

// Mode determinations behavior of the post-render manifest validation. Default is skip.
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

// Validator top level interfaced used by multirender to perform validation of all output files after the render has completed.
type Validator interface {
	dirValidator
	GetMode() Mode
}

// dirValidator is a package private interface that decouples the underlying mechanism performing validation from the options governing its behavior.
// The main advantage of this interface is that it would enable us to swap out kubeconform with any other validation engine without changing any code in package render.
// as long as the new mechanism implements this interface
type dirValidator interface {
	ValidateDir(path string) error
}

// validator providers a concrete implementation of the Validator interface
type validator struct {
	dirValidator
	Mode Mode
}

func (v validator) GetMode() Mode {
	return v.Mode
}

// New returns a new instance of a kubeconform Validator
func New(mode Mode, runner shell.Runner) Validator {
	return validator{kubeconform.New(runner), mode}
}

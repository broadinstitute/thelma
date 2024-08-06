package cli

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// errorRecorder records errors that take place during ThelmaCommand execution and returns a RunError
type errorRecorder struct {
	err   *RunError
	count int
}

func newErrorRecorder(key commandKey) *errorRecorder {
	return &errorRecorder{
		err: &RunError{
			CommandName: key.description(),
		},
		count: 0,
	}
}

func (r *errorRecorder) error() error {
	if !r.hasErrors() {
		return nil
	}
	return r.err
}

func (r *errorRecorder) hasErrors() bool {
	return r.count > 0
}

func (r *errorRecorder) addSetFlagsFromEnvironmentError(key commandKey, errs ...error) {
	for _, err := range errs {
		r.recordError(key, setFlagsFromEnvironment, err)
	}
}

func (r *errorRecorder) setRunError(key commandKey, err error) {
	r.recordError(key, runHook, err)
}

func (r *errorRecorder) setPreRunError(key commandKey, err error) {
	r.recordError(key, preRunHook, err)
}

func (r *errorRecorder) addPostRunError(key commandKey, err error) {
	r.recordError(key, postRunHook, err)
}

func (r *errorRecorder) recordError(key commandKey, hookType hookType, err error) {
	hookErr := &HookError{
		CommandName: key.description(),
		Err:         err,
		hookType:    hookType,
	}

	// Log error at debug level as it is recorded
	log.Debug().Str("phase", hookType.String()).Str("command", key.description()).Err(err).Msgf("error executing hook: %s", err.Error())

	switch hookType {
	case setFlagsFromEnvironment:
		r.err.SetFlagsFromEnvironmentErrors = append(r.err.SetFlagsFromEnvironmentErrors, hookErr)
	case preRunHook:
		r.err.PreRunError = hookErr
	case runHook:
		r.err.RunError = hookErr
	case postRunHook:
		r.err.PostRunErrors = append(r.err.PostRunErrors, hookErr)
	default:
		panic(errors.Errorf("Unknown hook type: %v", err))
	}

	r.count++
}

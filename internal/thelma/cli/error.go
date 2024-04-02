package cli

import (
	"fmt"
	"strings"
)

// RunError aggregates errors that can occur during execution of a ThelmaCommand
type RunError struct {
	CommandName                   string       // CommandName key of the Thelma command that was executed
	SetFlagsFromEnvironmentErrors []*HookError // SetFlagsFromEnvironmentErrors errors that occurred while setting flags from environment
	PreRunError                   *HookError   // PreRunError error returned by PreRun hook
	RunError                      *HookError   // RunError error returned by Run hook
	PostRunErrors                 []*HookError // PostRunErrors all errors returned by PostRun hooks
	Count                         int          // Count number of errors wrapped by this error
}

// Error returns a summary of all errors that occurred during command execution
func (e *RunError) Error() string {
	if e.Count <= 1 {
		// in single-error situations, just return the underlying error.
		// no need to clutter the message with extra context
		return e.firstError().Err.Error()
	}

	// build a summary of all errors
	var lines []string

	lines = append(lines, fmt.Sprintf("execution of command %q generated %d errors:", e.CommandName, e.Count))
	for _, setFlagsFromEnvironmentError := range e.SetFlagsFromEnvironmentErrors {
		lines = append(lines, setFlagsFromEnvironmentError.Error())
	}
	if e.PreRunError != nil {
		lines = append(lines, e.PreRunError.Error())
	}
	if e.RunError != nil {
		lines = append(lines, e.RunError.Error())
	}
	for _, postRunError := range e.PostRunErrors {
		lines = append(lines, postRunError.Error())
	}

	return strings.Join(lines, "\n")
}

// Cause implement the pkg/errors causer interface.
func (e *RunError) Cause() error {
	return e.firstError()
}

// firstError return the first error encountered during execution
func (e *RunError) firstError() *HookError {
	if len(e.SetFlagsFromEnvironmentErrors) > 0 {
		return e.SetFlagsFromEnvironmentErrors[0]
	}
	if e.RunError != nil {
		return e.RunError
	}
	if e.PreRunError != nil {
		return e.PreRunError
	}
	if len(e.PostRunErrors) > 0 {
		return e.PostRunErrors[0]
	}
	return nil
}

// HookError is an error returned from the execution of a single ThelmaCommand hook
type HookError struct {
	CommandName string   // CommandName key of the Thelma command the hook belongs to
	Err         error    // Err underlying error returned by the hook
	hookType    hookType // hookType type of hook that returned the error
}

func (e *HookError) Error() string {
	// generate message like `charts import (pre-run): error doing sthg`
	return fmt.Sprintf("%s (%s): %v", e.CommandName, e.hookType, e.Err)
}

// Cause implement the pkg/errors causer interface for stacktraces
func (e *HookError) Cause() error {
	return e.Err
}

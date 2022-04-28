package cli

import (
	"github.com/spf13/cobra"
)

// RunContext provides information about the current Thelma execution to PreRun, Run, and PostRun hooks
type RunContext interface {
	// Args returns user-supplied positional arguments for the leaf command.
	Args() []string

	// CobraCommand returns the Cobra command associated with the Thelma command that is currently executing.
	CobraCommand() *cobra.Command

	// Parent returns the parent of the Thelma command that is currently executing.
	Parent() ThelmaCommand

	// SetOutput sets the output for this command (will be converted to YAML, JSON, dump, or raw based on user-supplied arguments, and printed to stdout)
	SetOutput(data interface{})

	// HasOutput returns true if output has been set for this command
	HasOutput() bool

	// Output returns output for this command, or nil if none has been set
	Output() interface{}

	// CommandName returns the name components for this command. Eg. ["render"] for `thelma render`, ["bee", "list"] for `thelma bee list`
	CommandName() []string
}

func newRunContext(key commandKey, args []string) *runContext {
	return &runContext{
		commandKey: key,
		args:       args,
	}
}

// runContext implements the RunContext interface (see api package)
type runContext struct {
	commandKey             commandKey
	args                   []string
	currentlyExecutingNode *node
	output                 interface{}
	hasOutput              bool
}

// sets the currently executing node
func (r *runContext) setCurrentExecutingNode(node *node) {
	r.currentlyExecutingNode = node
}

func (r *runContext) CobraCommand() *cobra.Command {
	return r.currentlyExecutingNode.cobraCommand
}

func (r *runContext) Args() []string {
	return r.args
}

func (r *runContext) Parent() ThelmaCommand {
	if r.currentlyExecutingNode.isRoot() {
		return nil
	}
	return r.currentlyExecutingNode.parent.thelmaCommand
}

func (r *runContext) SetOutput(data interface{}) {
	r.output = data
	r.hasOutput = true
}

func (r *runContext) HasOutput() bool {
	return r.hasOutput
}

func (r *runContext) Output() interface{} {
	return r.output
}

func (r *runContext) CommandName() []string {
	return r.commandKey.nameComponents
}

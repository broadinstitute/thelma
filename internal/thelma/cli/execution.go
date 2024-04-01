package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics"
	"github.com/broadinstitute/thelma/internal/thelma/app/version"
	"github.com/broadinstitute/thelma/internal/thelma/cli/environmentflags"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

// executionOptions options for an execution
type executionOptions struct {
	skipRun bool // skipRun if true, execute preRun/postRun hooks but NOT the main run hook
}

// execution is for executing Thelma commands.
type execution struct {
	options       *executionOptions
	leafNode      *node
	app           app.ThelmaApp
	runContext    *runContext
	errorRecorder *errorRecorder
	startTime     time.Time
}

// newExecution is a constructor for an execution
func newExecution(options *executionOptions, leafNode *node, builder builder.ThelmaBuilder, args []string) (*execution, error) {
	_app, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return &execution{
		options:       options,
		leafNode:      leafNode,
		app:           _app,
		runContext:    newRunContext(leafNode.key, args),
		errorRecorder: newErrorRecorder(leafNode.key),
	}, nil
}

// execute runs a Thelma command, including pre and post run hooks.
// all errors are aggregated into a RunError
func (e *execution) execute() error {
	e.startTime = time.Now()
	e.setFlagsFromEnvironment()
	e.preRun()
	e.run()
	e.postRun()
	e.recordExecutionMetrics()

	closeErr := e.app.Close()
	if closeErr != nil {
		log.Warn().Err(closeErr).Msgf("error cleaning up thelma: %v", closeErr)
	}

	return e.errorRecorder.error()
}

// setFlagsFromEnvironment calls out to the environmentflags.SetFlagsFromEnvironment. This function does two things:
//
//	1, it uses e.errorRecorder to handle any errors so we clean up properly
//	2, it encapsulates punching a hole in the normal node order
//
// That second point is the most subtle and most important. See ThelmaCommand.PreRun, ThelmaCommand.Run, and
// ThelmaCommand.PostRun for context on the order that each node's hooks are executed. That order is perfect for
// most things, but not loading flags from the environment.
//
// To set flags from the environment, we need access to the full set of flags for the command. That means we actually
// want to use the flags that cobra.Command makes visible on the execution.leafNode (in other words, what's shown in
// `thelma <command> --help`). If we were to access the flags from any other node, we'd be getting the set of flags
// visible if that node's command was the "leaf" being executed.
//
// Our first instinct might be to add environment parsing to the leafNode's ThelmaCommand.PreRun hook. The problem is
// that parent nodes' ThelmaCommand.PreRun would've executed first! If a parent node set a persistent flag, the parent
// node could've already read it in its own ThelmaCommand.PreRun before we get a chance to parse from the environment.
// We need to grab the leaf node and handle environment parsing before *any* hooks. That way, every hook will see the
// same values for the flags visible to each node.
func (e *execution) setFlagsFromEnvironment() {
	if errs := environmentflags.SetFlagsFromEnvironment(e.leafNode.cobraCommand.Flags()); len(errs) > 0 {
		e.errorRecorder.addSetFlagsFromEnvironmentError(e.leafNode.key, errs...)
	}
}

// preRun executes pre-run phase of a Thelma command execution
func (e *execution) preRun() {
	// Run ancestor and leafNode PreRun hooks in descending order (root first)
	for _, n := range pathFromRoot(e.leafNode) {
		e.runContext.setCurrentExecutingNode(n)
		if err := n.thelmaCommand.PreRun(e.app, e.runContext); err != nil {
			e.errorRecorder.setPreRunError(n.key, err)
			// stop executing in event of an error
			return
		}
	}
}

// run executes the run phase of a Thelma command execution
func (e *execution) run() {
	if e.options.skipRun {
		log.Debug().Msgf("skipping run phase, hookOnly=%v", e.options.skipRun)
		return
	}
	if e.errorRecorder.hasErrors() {
		log.Debug().Msgf("skipping run phase, pre-run returned an error")
		return
	}
	e.runContext.setCurrentExecutingNode(e.leafNode)
	if err := e.leafNode.thelmaCommand.Run(e.app, e.runContext); err != nil {
		e.errorRecorder.setRunError(e.leafNode.key, err)
	}
}

// postRun executes post-run phase of a Thelma command execution
func (e *execution) postRun() {
	// Run ancestor and leafNode PreRun hooks in ascending order (root last)
	for _, n := range pathToRoot(e.leafNode) {
		e.runContext.setCurrentExecutingNode(n)
		if err := n.thelmaCommand.PostRun(e.app, e.runContext); err != nil {
			e.errorRecorder.addPostRunError(n.key, err)
			// keep executing in event of an error. (post-run hooks are always guaranteed to run)
		}
	}
}

func (e *execution) recordExecutionMetrics() {
	err := e.errorRecorder.error()
	duration := time.Since(e.startTime)

	opts := metrics.Options{
		Name: "run",
		Help: "Thelma command run",
		Labels: map[string]string{
			"command": strings.Join(e.runContext.commandKey.nameComponents, "_"),
		},
	}

	metrics.TaskCompletion(opts, duration, err)

	metrics.Counter(metrics.Options{
		Name: "version",
		Help: "Records the version of Thelma used for a given Thelma run",
		Labels: map[string]string{
			"version": version.Version,
		},
	}).Inc()
}

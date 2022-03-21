// Package cli contains code for Thelma's command-line interface
package cli

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/spf13/cobra"
)

type ThelmaCLI interface {
	// Execute is the main entry point for Thelma execution. It can only be called once for a given ThelmaCLI.
	Execute() error
}

// implements the ThelmaCLI interface
type thelmaCLI struct {
	treeRoot *node
}

func (t *thelmaCLI) Execute() error {
	return t.treeRoot.cobraCommand.Execute()
}

func New(options ...Option) ThelmaCLI {
	asStruct := DefaultOptions()
	for _, option := range options {
		option(asStruct)
	}
	return NewWithOptions(asStruct)
}

func NewWithOptions(options *Options) ThelmaCLI {
	// create thelma builder
	thelmaBuilder := builder.NewBuilder()

	// create root command
	root := newTree(options.commands)

	eopts := &executionOptions{
		skipRun: options.skipRun,
	}

	// now configure cobra for every node in the tree
	preOrderTraverse(root, func(n *node) {
		// set useful defaults on the Cobra command
		n.cobraCommand.Use = fmt.Sprintf("%s [options]", n.key.shortName())

		// add RunE function for leaf Cobra commands (intermediate/non-leaf commands just print out help messages)
		if n.isLeaf() {
			n.cobraCommand.RunE = func(_ *cobra.Command, args []string) error {
				e, err := newExecution(eopts, n, thelmaBuilder, args)
				if err != nil {
					return err
				}
				return e.execute()
			}
		}

		// run user-supplied configure hook to add flags, description, etc to the command
		n.thelmaCommand.ConfigureCobra(n.cobraCommand)

		// add Cobra command as a child of its parent command
		if !n.isRoot() {
			n.parent.cobraCommand.AddCommand(n.cobraCommand)
		}
	})

	// handle options
	if options.thelmaConfigHook != nil {
		options.thelmaConfigHook(thelmaBuilder)
	}
	for _, cobraHook := range options.cobraConfigHooks {
		cobraHook(root.cobraCommand)
	}

	return &thelmaCLI{
		treeRoot: root,
	}
}

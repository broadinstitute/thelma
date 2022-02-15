package cli

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/spf13/cobra"
)

// ThelmaCommand must be implemented by all Thelma subcommands.
type ThelmaCommand interface {
	// ConfigureCobra is a hook for adding flags, description, and other configuration to the Cobra command associated with this Thelma subcommand
	ConfigureCobra(cobraCommand *cobra.Command)

	// PreRun is a hook for running some code before Run is called. Conventionally, flag and argumnent validation are
	// performed here. Note that:
	// * PreRun hooks are inherited by child commands, and run in order of inheritance. Eg.
	//   for the command "thelma charts import", the root command's PreRun will be run first,
	//   then the `chart` command's PreRun, and finally the `import` command's PreRun.
	//   If an error occurs in a prent PreRun hook, it will be returned; child/descendant PreRun hooks will not be executed.
	// * If any PreRun hook returns an error, the command's Run hook will not be executed, but its PostRun hooks will be.
	PreRun(app app.ThelmaApp, ctx RunContext) error

	// Run is where the main body of a subcommand should be implemented. Note that:
	// * For commands with subcommands (eg. the "charts" command in "thelma charts import"), the Run hook is ignored.
	// * If PreRun returns an error, Run is not called but PostRun is.
	Run(app app.ThelmaApp, ctx RunContext) error

	// PostRun is a hook for running some code after Run is called. Conventionally, cleanup goes here.
	// Note that:
	// * PostRun hooks are inherited by child commands, and run in reverse order of inheritance. Eg.
	//	 for the command "thelma charts import", the `import` command's PostRun will be run first,
	//	 then the `chart` command's PostRun, and finally the root command's PostRun.
	// * Unlike PreRun hooks, PostRun hooks are guaranteed to run, even if an earlier PostRun, PreRun, or Run hook fails.
	// * PostRun hooks should be written carefully to avoid errors in the event of an earlier failure (check pointers for nil, etc).
	PostRun(app app.ThelmaApp, ctx RunContext) error
}

# cli

This package implements the command-line interface for Thelma. It provides some additional functionality on top of [Cobra](https://github.com/spf13/cobra).

### Why?

Cobra's awesome, but it has a few missing features.

#### Dependency Injection

For tools with many subcommands, it's likely that most commands will share some central dependencies (eg. configuration, logging, API clients). Cobra provides no mechanism for injecting dependencies into subcommands; users must roll their own DI in a way that is adaptable to testing.

#### Accessing Options Structs in Tests

Most Cobra commands use BindFlags to associate CLI flags with a custom options struct. It's useful to be able to access and verify these structs in tests, especially for commands that have complex options, but Cobra provides no mechanism for this.

#### Guaranteed Hook Execution

Cobra's PreRun and PostRun hook implementation are missing some key features:

* **Hooks are not inherited** by child commands. (PersistentPreRun and PersistentPostRun are inherited, but not transitively). In an ideal world, we would be able to configure pre- and post- hooks on the root command, an arbitrary set of intermediate commands, and finally the leaf/child command, and have all hooks be executed in order.

* **PostRun hooks won't be executed** if the main Run() hook fails. This is limiting, because often it is useful to do some initialization in a PreRun hook and cleanup in a corresponding PostRun hook.
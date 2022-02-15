package cli

import (
	"fmt"
	"regexp"
	"strings"
)

const reservedRootCommandName = "root"

var validCommandName = regexp.MustCompile("^[a-z0-9-]+$")

// commandKey is represents a unique key for a Thelma command in the command tree, derived from the command name
type commandKey struct {
	// nameComponents slice of command key components, excluding the `thelma` part. Eg.
	// `thelma render` ->  {"render"}
	// `thelma charts import` - {"charts", "import"}
	// the root command is represented as an empty slice []string{}
	nameComponents []string
}

// newCommandKey constructor for commandKey
func newCommandKey(name string) commandKey {
	return commandKey{nameComponents: strings.Fields(name)}
}

// rootCommandKey returns the key for the root command
func rootCommandKey() commandKey {
	return commandKey{nameComponents: []string{}}
}

// validateCommandName returns an error if a command key is invalid
func validateCommandName(fullName string) error {
	components := strings.Fields(fullName)
	for _, component := range components {
		if component == reservedRootCommandName {
			return fmt.Errorf("invalid command key (%q is reserved): %v", reservedRootCommandName, fullName)
		}
		if !validCommandName.MatchString(component) {
			return fmt.Errorf("invalid command key (each component must match %q): %v", validCommandName.String(), fullName)
		}
	}
	return nil
}

// description returns a string description of the command, suitable for use in log and error messages. Identical
// to longName except "root" is returned for the root command instead of the empty string.
// eg. [] (root) -> "root"
//     ["render"] -> "render"
//     ["charts", "import"] -> "charts import"
func (n commandKey) description() string {
	if n.isRoot() {
		return reservedRootCommandName
	}
	return strings.Join(n.nameComponents, " ")
}

// depth returns nest level for this command
// eg. [] (root) -> 0
//     ["render"] -> 1,
//     ["charts", "import"] -> 2
func (n commandKey) depth() int {
	return len(n.nameComponents)
}

// longName returns unique long key for this command, including ancestors. Suitable for use as a hash key.
// eg. [] (root) -> ""
//     ["render"] -> "render"
//     ["charts", "import"] -> "charts import"
//     ["data", "import"] -> "data import"
func (n commandKey) longName() string {
	return strings.Join(n.nameComponents, " ")
}

// shortName returns short / leaf key of this command, without ancestors.
// eg. [] (root) -> ""
//     ["render"] -> "render"
//     ["charts", "import"] -> "import"
//     ["bee", "import"] -> "import"
func (n commandKey) shortName() string {
	if n.isRoot() { // avoid out of bounds for root command
		return ""
	}
	return n.nameComponents[len(n.nameComponents)-1]
}

// isRoot returns true if this command key is empty, i.e. the key of the root command
func (n commandKey) isRoot() bool {
	return len(n.nameComponents) == 0
}

// ancestors returns ancestor components of the command name
// eg. ["render"] -> []
//     ["charts", "import"] -> ["charts"]
//     [] (root) -> []
//     ["a", "b", "c"] -> ["a", "b"]
func (n commandKey) ancestors() []string {
	if n.isRoot() {
		return []string{}
	}
	return n.nameComponents[0 : len(n.nameComponents)-1]
}

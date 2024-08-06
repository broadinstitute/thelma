package environmentflags

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

const flagsFromEnvironmentPrefixFlag = "flags-from-environment-prefix"

// AddFlag adds the magic flag that enables the environment variable parsing feature.
// See SetFlagsFromEnvironment for more information.
func AddFlag(flags *pflag.FlagSet) {
	flags.String(flagsFromEnvironmentPrefixFlag, "",
		`Optionally read other flags from environment (if set to "PARAM_", "PARAM_OUTPUT_FORMAT=json" would set --output-format)`)
}

// SetFlagsFromEnvironment is enabled/configured by a user setting the magic flag in
// `flagsFromEnvironmentPrefixFlag`. If set, this function looks for the value of each
// *other* flag in the environment, and updates the given flags as if any found values
// we passed on the command line.
//
// For example, if the magic flag was set to "PARAM_", and the flags contained a flag
// named "foo-bar", this function would look for a variable named "PARAM_FOO_BAR".
// There is some type coercion borrowed from the Cobra CLI framework, so if "foo-bar"
// was a boolean flag, it could be set with "PARAM_FOO_BAR=true", "PARAM_FOO_BAR=1",
// and so on.
//
// This function can't help set positional arguments, it only helps with flags.
//
// This function largely exists so that tools like ArgoCD that want to run Thelma
// with configuration in the environment can do so without needing a wrapper script.
func SetFlagsFromEnvironment(flags *pflag.FlagSet) []error {
	// We only read flags from environment if the configuration flag was passed.
	if !flags.Changed(flagsFromEnvironmentPrefixFlag) {
		return nil
	}

	prefix, err := flags.GetString(flagsFromEnvironmentPrefixFlag)
	if err != nil {
		return []error{errors.Errorf("--%s must be a string if set", flagsFromEnvironmentPrefixFlag)}
	}

	// Iterate over each flag. We must accumulate errors, since we can't return early.
	var errs []error
	flags.VisitAll(func(flag *pflag.Flag) {

		// If the flag isn't our own config flag, and it wasn't set on the command line,
		// we'll try to read it.
		if flag.Name != flagsFromEnvironmentPrefixFlag && !flag.Changed {

			// We look for the prefix immediately followed by the upper-snake-case form
			// of the flag name.
			envVar := prefix + strings.ToUpper(strings.ReplaceAll(flag.Name, "-", "_"))

			// If the environment variable is set, even if it's empty, we'll set the flag.
			if value, present := os.LookupEnv(envVar); present {
				// We use flags.Set instead of flag.Value.Set because the latter doesn't set
				// the flag as "changed", which is important for how we validate inputs.
				if flagSetErr := flags.Set(flag.Name, value); flagSetErr != nil {
					// Include the value in the error message for better debugging; this is
					// safe enough because we don't accept secret values from flags.
					errs = append(errs, errors.Wrapf(flagSetErr,
						"failed to set --%s from environment variable %s with value `%s`",
						flag.Name, envVar, value))
				}
			}
		}
	})

	// Either there's some errors or there's none, in either case we just return the list.
	return errs
}

package selector

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

// valueSet represents user-supplied values to an enum flag (either a set of distinct values or an ALL selector)
type valueSet struct {
	isAllSelector bool
	uniqueValues  []string
}

// enumFlag is for processing cli flags that can contain either
//  (1) a set of possible values (eg. "--environments foo,bar,baz")
//  (2) an ALL selector (meaning include all values, eg. "--environments ALL")
// and produce a filter
type enumFlag struct {
	// flagName long name of this flag, eg. "releases", "environments", etc
	flagName string

	// shortHand single-letter alias for the flag. Can be left blank for no short-hand
	shortHand string

	// usageMessage usage message for this flag
	usageMessage string

	// defaultValues default values for this flag
	defaultValues []string

	// userValues holds the values supplied by the user to the flag argument. Populated by Cobra
	userValues []string

	// preProcessHook callback function that can be used to manipulate user-supplied flag values before further processing
	preProcessHook func(flagValues []string, args []string, changed bool) (normalizedValues []string, err error)

	// validValues callback function that must return the set of valid values for this flag
	validValues func(state terra.State) (set.StringSet, error)

	// buildFilter callback function that builds a filter and adds it to the filterBuilder
	buildFilter func(f *filterBuilder, uniqueValues []string)
}

// addToCobraCommand registers this flag with a cobra command
func (e *enumFlag) addToCobraCommand(cobraCommand *cobra.Command) {
	if e.shortHand == "" {
		cobraCommand.Flags().StringSliceVar(&e.userValues, e.flagName, e.defaultValues, e.usageMessage)
	} else {
		cobraCommand.Flags().StringSliceVarP(&e.userValues, e.flagName, e.shortHand, e.defaultValues, e.usageMessage)
	}
}

// processInput convert user input into a filter and add it to the filterBuilder
func (e *enumFlag) processInput(f *filterBuilder, state terra.State, changed bool, args []string) error {
	inputValues := e.userValues

	if e.preProcessHook != nil {
		var err error
		inputValues, err = e.preProcessHook(e.userValues, args, changed)
		if err != nil {
			return err
		}
	}

	vset, err := e.validate(state, inputValues)
	if err != nil {
		return err
	}
	if vset.isAllSelector {
		// no need to filter if ALL selector was supplied
		return nil
	}

	e.buildFilter(f, vset.uniqueValues)
	return nil
}

// validate validates user-supplied flag values, returning an valueSet struct describing the values
func (e *enumFlag) validate(state terra.State, inputValues []string) (*valueSet, error) {
	collated := collateSelectorValues(inputValues)

	if collated.Empty() {
		return nil, fmt.Errorf("--%s: at least option must be speficied", e.flagName)
	}

	// handle ALL selector (--releases=ALL, --cluster=ALL, etc)
	if collated.Exists(allSelector) {
		if collated.Size() > 1 {
			return nil, fmt.Errorf("--%s: either %s or individual %s can be specified, but not both: %s", e.flagName, allSelector, e.flagName, strings.Join(collated.Elements(), ", "))
		}
		return &valueSet{
			isAllSelector: true,
			uniqueValues:  collated.Elements(),
		}, nil
	}

	// make sure enumerated values are valid
	if e.validValues != nil {
		valid, err := e.validValues(state)
		if err != nil {
			return nil, fmt.Errorf("error identifying valid values for flag --%s: %v", e.flagName, err)
		}

		diff := collated.Difference(valid)
		if !diff.Empty() {
			unknown := diff.Elements()
			sort.Strings(unknown)
			return nil, fmt.Errorf("--%s: unknown %s %s", e.flagName, maybePlural(e.flagName), strings.Join(unknown, ", "))
		}
	}

	return &valueSet{
		isAllSelector: false,
		uniqueValues:  collated.Elements(),
	}, nil
}

// collateSelectorValues collects selector values into a string set
// given a slice of strings like
//   "foo,bar", "bar foo", "baz", "quux,baz"
// return a set consisting of the elements:
//   "foo", "bar", "baz", "quux"
func collateSelectorValues(values []string) set.StringSet {
	_set := set.NewStringSet()

	for _, value := range values {
		for _, token := range strings.Split(value, selectorSeparator) {
			token = strings.TrimSpace(token)
			if token != "" {
				_set.Add(token)
			}
		}
	}

	return _set
}

// "environment" -> "environment(s)"
// "cluster" -> "cluster(s)"
func maybePlural(singular string) string {
	return fmt.Sprintf("%s(s)", singular)
}

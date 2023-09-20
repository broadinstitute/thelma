package filterflags

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli/flags"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"time"
)

type flagValues struct {
	template     string
	nameIncludes string
	olderThan    time.Duration
}

var flagNames = struct {
	template     string
	nameIncludes string
	olderThan    string
}{
	template:     "template",
	nameIncludes: "name-includes",
	olderThan:    "older-than",
}

type filterFlags struct {
	options  flags.Options
	flagVals flagValues
}

// FilterFlags adds bee filtering flags to a cobra command and supports converting those flags to a terra.EnvironmentFilter
type FilterFlags interface {
	// AddFlags add bee filtering flags such as --name-includes, --older-than, and --template to a command
	AddFlags(*cobra.Command)
	// GetFilter should be called during a Run function to get a terra.EnvironemntFilter that matches the given filter flags
	GetFilter(thelmaApp app.ThelmaApp) (terra.EnvironmentFilter, error)
}

// NewFilterFlags returns a new filterFlags
func NewFilterFlags(opts ...flags.Option) FilterFlags {
	return &filterFlags{
		options: flags.AsOptions(opts),
	}
}

func (f *filterFlags) AddFlags(cobraCommand *cobra.Command) {
	f.options.Apply(cobraCommand.Flags(), func(flags *pflag.FlagSet) {
		flags.StringVarP(&f.flagVals.template, flagNames.template, "t", "", "Only include BEEs created from the given template")
		flags.StringVarP(&f.flagVals.nameIncludes, flagNames.nameIncludes, "i", "", "Only include BEEs with names that include the given substring")
		flags.DurationVar(&f.flagVals.olderThan, flagNames.olderThan, 0, "Only include BEEs older than the given duration")
	})
}

func (f *filterFlags) GetFilter(thelmaApp app.ThelmaApp) (terra.EnvironmentFilter, error) {
	var filters []terra.EnvironmentFilter

	state, err := thelmaApp.State()
	if err != nil {
		return nil, err
	}

	if f.flagVals.template != "" {
		template, err := state.Environments().Get(f.flagVals.template)
		if err != nil {
			return nil, err
		}
		if template == nil {
			return nil, errors.Errorf("--%s: no template by the name %q exists", flagNames.template, f.flagVals.template)
		}
		filters = append(filters, filter.Environments().HasTemplate(template))
	}

	if f.flagVals.nameIncludes != "" {
		filters = append(filters, filter.Environments().NameIncludes(f.flagVals.nameIncludes))
	}

	if f.flagVals.olderThan > 0 {
		filters = append(filters, filter.Environments().OlderThan(f.flagVals.olderThan))
	}

	return filter.Environments().And(filters...), nil
}

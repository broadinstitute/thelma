package filter

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"strings"
	"time"
)

func Environments() EnvironmentFilters {
	return environmentFilters{}
}

// EnvironmentFilters provides helper functions for building terra.EnvironmentFilter instances
type EnvironmentFilters interface {
	// Any returns a filter that matches any environment
	Any() terra.EnvironmentFilter
	// HasBase returns a filter that matches environments with the given configuration base(s)
	HasBase(base ...string) terra.EnvironmentFilter
	// HasLifecycle matches environments with the given lifecycle(s)
	HasLifecycle(lifecycle terra.Lifecycle) terra.EnvironmentFilter
	// HasLifecycleName matches environments with the given lifecycle name(s)
	HasLifecycleName(lifecycleName ...string) terra.EnvironmentFilter
	// HasTemplate matches environments with the given template
	HasTemplate(template terra.Environment) terra.EnvironmentFilter
	// HasTemplateName matches environments with the given template name(s)
	HasTemplateName(templateNames ...string) terra.EnvironmentFilter
	// NameIncludes returns environments with names that include the given substring
	NameIncludes(substring string) terra.EnvironmentFilter
	// OlderThan returns environments that are older than a given duration
	OlderThan(dur time.Duration) terra.EnvironmentFilter
	// AutoDeletable returns environments that can be automatically deleted
	AutoDeletable() terra.EnvironmentFilter
	// Or returns a filter that matches environments that match _any_ of the given filters
	Or(filters ...terra.EnvironmentFilter) terra.EnvironmentFilter
	//And returns a filter that matches environments that match _all_ of the given filters
	And(filters ...terra.EnvironmentFilter) terra.EnvironmentFilter
}

// implements the EnvironmentFilters interface
type environmentFilters struct{}

func (e environmentFilters) Any() terra.EnvironmentFilter {
	return environmentFilter{
		string: anyString,
		matcher: func(_ terra.Environment) bool {
			return true
		},
	}
}

func (e environmentFilters) HasBase(bases ...string) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf("hasBase(%s)", join(quote(bases)...)),
		matcher: func(environment terra.Environment) bool {
			for _, base := range bases {
				if environment.Base() == base {
					return true
				}
			}
			return false
		},
	}
}

func (e environmentFilters) HasLifecycle(lifecycle terra.Lifecycle) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf("hasLifecycle(%s)", lifecycle.String()),
		matcher: func(environment terra.Environment) bool {
			return environment.Lifecycle() == lifecycle
		},
	}
}

func (e environmentFilters) HasLifecycleName(lifecycleNames ...string) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf("hasLifecycleName(%s)", join(quote(lifecycleNames)...)),
		matcher: func(environment terra.Environment) bool {
			for _, lifecycleName := range lifecycleNames {
				if environment.Lifecycle().String() == lifecycleName {
					return true
				}
			}
			return false
		},
	}
}

func (e environmentFilters) HasTemplate(template terra.Environment) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf("hasTemplate(%s)", template.Name()),
		matcher: func(environment terra.Environment) bool {
			return environment.Template() == template.Name()
		},
	}
}

func (e environmentFilters) HasTemplateName(templateNames ...string) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf("hasTemplateNames(%s)", join(quote(templateNames)...)),
		matcher: func(environment terra.Environment) bool {
			if !environment.Lifecycle().IsDynamic() {
				// only dynamic environments have templates
				return false
			}
			for _, t := range templateNames {
				if environment.Template() == t {
					return true
				}
			}
			return false
		},
	}
}

func (e environmentFilters) NameIncludes(substring string) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf("nameIncludes(%q)", substring),
		matcher: func(environment terra.Environment) bool {
			return strings.Contains(environment.Name(), substring)
		},
	}
}

func (e environmentFilters) OlderThan(dur time.Duration) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf("olderThan(%s)", dur),
		matcher: func(environment terra.Environment) bool {
			cutoffTime := time.Now().Add(-dur)
			return environment.CreatedAt().Before(cutoffTime)
		},
	}
}

func (e environmentFilters) AutoDeletable() terra.EnvironmentFilter {
	return environmentFilter{
		string: "autoDeletable()",
		matcher: func(environment terra.Environment) bool {
			return environment.Lifecycle() == terra.Dynamic &&
				!environment.PreventDeletion() &&
				environment.AutoDelete().Enabled() &&
				environment.AutoDelete().After().Before(time.Now())
		},
	}
}

//
// TODO [generics] Or and And functions are duplicated across all filter types, fix when generics are available

func (e environmentFilters) Or(filters ...terra.EnvironmentFilter) terra.EnvironmentFilter {
	if len(filters) == 0 {
		return e.Any()
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return environmentFilter{
		string: fmt.Sprintf(orFormat, join(environmentFilterStrings(filters)...)),
		matcher: func(env terra.Environment) bool {
			for _, f := range filters {
				if f.Matches(env) {
					return true
				}
			}
			return false
		},
	}
}

func (e environmentFilters) And(filters ...terra.EnvironmentFilter) terra.EnvironmentFilter {
	if len(filters) == 0 {
		return e.Any()
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return environmentFilter{
		string: fmt.Sprintf(andFormat, join(environmentFilterStrings(filters)...)),
		matcher: func(env terra.Environment) bool {
			for _, f := range filters {
				if !f.Matches(env) {
					return false
				}
			}
			return true
		},
	}
}

func environmentFilterStrings(filters []terra.EnvironmentFilter) []string {
	var filterStrings []string
	for _, f := range filters {
		filterStrings = append(filterStrings, f.String())
	}
	return filterStrings
}

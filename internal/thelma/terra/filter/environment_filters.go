package filter

import "github.com/broadinstitute/thelma/internal/thelma/terra"

func Environments() EnvironmentFilters {
	return environmentFilters{}
}

// EnvironmentFilters provides helper functions for building terra.EnvironmentFilter instances
type EnvironmentFilters interface {
	// Any returns a filter that matches any environment
	Any() terra.EnvironmentFilter
	// HasLifecycle matches environments with the given lifecyle
	HasLifecycle(terra.Lifecycle) terra.EnvironmentFilter
	// HasTemplate matches environments with the given template
	HasTemplate(environment terra.Environment) terra.EnvironmentFilter
}

// implements the EnvironmentFilters interface
type environmentFilters struct{}

func (e environmentFilters) Any() terra.EnvironmentFilter {
	return environmentFilter{
		matcher: func(_ terra.Environment) bool {
			return true
		},
	}
}

func (e environmentFilters) HasLifecycle(lifecycle terra.Lifecycle) terra.EnvironmentFilter {
	return environmentFilter{
		matcher: func(environment terra.Environment) bool {
			return environment.Lifecycle() == lifecycle
		},
	}
}

func (e environmentFilters) HasTemplate(template terra.Environment) terra.EnvironmentFilter {
	return environmentFilter{
		matcher: func(environment terra.Environment) bool {
			return environment.Template() == template.Name()
		},
	}
}

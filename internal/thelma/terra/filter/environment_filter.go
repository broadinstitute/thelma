package filter

import "github.com/broadinstitute/thelma/internal/thelma/terra"

// TODO replace separate environment, release, and destination filter implementations with generics once they're available

// implements the terra.EnvironmentFilter interface
type environmentFilter struct {
	matcher func(terra.Environment) bool
}

func (f environmentFilter) Matches(environment terra.Environment) bool {
	return f.matcher(environment)
}

func (f environmentFilter) And(other terra.EnvironmentFilter) terra.EnvironmentFilter {
	return environmentFilter{
		matcher: func(environment terra.Environment) bool {
			return f.Matches(environment) && other.Matches(environment)
		},
	}
}

func (f environmentFilter) Or(other terra.EnvironmentFilter) terra.EnvironmentFilter {
	return environmentFilter{
		matcher: func(environment terra.Environment) bool {
			return f.Matches(environment) || other.Matches(environment)
		},
	}
}

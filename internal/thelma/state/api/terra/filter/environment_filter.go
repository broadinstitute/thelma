package filter

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

// TODO replace separate environment, release, and destination filter implementations with generics once they're available

// implements the terra.EnvironmentFilter interface
type environmentFilter struct {
	string  string
	matcher func(terra.Environment) bool
}

func (f environmentFilter) String() string {
	return f.string
}

func (f environmentFilter) Matches(environment terra.Environment) bool {
	return f.matcher(environment)
}

func (f environmentFilter) And(other terra.EnvironmentFilter) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf(andFormat, join(f.String(), other.String())),
		matcher: func(environment terra.Environment) bool {
			return f.Matches(environment) && other.Matches(environment)
		},
	}
}

func (f environmentFilter) Or(other terra.EnvironmentFilter) terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf(orFormat, join(f.String(), other.String())),
		matcher: func(environment terra.Environment) bool {
			return f.Matches(environment) || other.Matches(environment)
		},
	}
}

func (f environmentFilter) Negate() terra.EnvironmentFilter {
	return environmentFilter{
		string: fmt.Sprintf(notFormat, f.String()),
		matcher: func(env terra.Environment) bool {
			return !f.Matches(env)
		},
	}
}

func (f environmentFilter) Filter(environments []terra.Environment) []terra.Environment {
	var result []terra.Environment
	for _, environment := range environments {
		if f.Matches(environment) {
			result = append(result, environment)
		}
	}
	return result
}

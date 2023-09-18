package terra

// Generics can't come soon enough :'(

// A ReleaseFilter is a predicate for filtering lists of releases
type ReleaseFilter interface {
	// String returns a string representation of the filter
	String() string
	// Matches returns true if this filter matches the release
	Matches(Release) bool
	// And returns a new filter that matches this filter and another
	And(ReleaseFilter) ReleaseFilter
	// Or returns a new filter that matches this filter or another
	Or(ReleaseFilter) ReleaseFilter
	// Negate returns a new filter that matches the opposite of this filter
	Negate() ReleaseFilter
	// Filter given a list of releases, return the sublist that match this filter
	Filter([]Release) []Release
}

// A DestinationFilter is a predicate for filtering lists of destinations
type DestinationFilter interface {
	// String returns a string representation of the filter
	String() string
	// Matches returns true if this filter matches the destination
	Matches(Destination) bool
	// And returns a new filter that matches this filter and another
	And(DestinationFilter) DestinationFilter
	// Or returns a new filter that matches this filter or another
	Or(DestinationFilter) DestinationFilter
	// Filter given a list of destinations, return the sublist that match this filter
	Filter([]Destination) []Destination
}

// An EnvironmentFilter is a predicate for filtering lists of environments
type EnvironmentFilter interface {
	// String returns a string representation of the filter
	String() string
	// Matches returns true if this filter matches the environment
	Matches(Environment) bool
	// And returns a new filter that matches this filter and another
	And(EnvironmentFilter) EnvironmentFilter
	// Or returns a new filter that matches this filter or another
	Or(EnvironmentFilter) EnvironmentFilter
	// Filter given a list of environments, return the sublist that match this filter
	Filter([]Environment) []Environment
}

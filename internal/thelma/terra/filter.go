package terra

// Generics can't come soon enough :'(

// A ReleaseFilter is a predicate for filtering lists of releases
type ReleaseFilter interface {
	Matches(Release) bool
	And(ReleaseFilter) ReleaseFilter
	Or(ReleaseFilter) ReleaseFilter
}

// A DestinationFilter is a predicate for filtering lists of destinations
type DestinationFilter interface {
	Matches(Destination) bool
	And(DestinationFilter) DestinationFilter
	Or(DestinationFilter) DestinationFilter
}

// An EnvironmentFilter is a predicate for filtering lists of environments
type EnvironmentFilter interface {
	Matches(Environment) bool
	And(EnvironmentFilter) EnvironmentFilter
	Or(EnvironmentFilter) EnvironmentFilter
}

package terra

// Destinations is an interface for querying release destinations
type Destinations interface {
	// All returns a list of all destinations
	All() ([]Destination, error)
	// Filter returns a list of clusters matching the given filter
	Filter(filter DestinationFilter) ([]Destination, error)
	// Get returns the destination with the given name, or an error if no such destination exists
	Get(name string) (Destination, error)
}

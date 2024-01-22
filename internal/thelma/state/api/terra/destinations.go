package terra

// Destinations is an interface for querying release destinations
type Destinations interface {
	// All returns a list of all destinations
	All() ([]Destination, error)
	// Get returns the destination with the given name, or an error if no such destination exists
	Get(name string) (Destination, error)
}

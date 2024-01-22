package terra

// Releases is an interface for querying releases
type Releases interface {
	// All returns a list of all releases
	All() ([]Release, error)
}

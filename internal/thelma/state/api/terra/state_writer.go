package terra

// StateWriteris an interface for exporting thelma's internal state to some external location
// currently this is sherlock but could also be a file or something else in the future
type StateWriter interface {
	WriteClusters([]Cluster) error
	// WriteEnvironments writes environments into thelma's state source it will return either a list of
	// newly created environment names if successful or an error
	WriteEnvironments([]Environment) ([]string, error)
	// Delete Environments is used to delete dynamic environments from Thelma's state source.
	// It returns a list of the names of deleted environments
	DeleteEnvironments([]Environment) ([]string, error)

	EnableRelease(Environment, string) error
	DisableRelease(string, string) error
}

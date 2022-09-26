package terra

// StateExporter is an interface for exporting thelma's internal state to some external location
// currently this is sherlock but could also be a file or something else in the future
type StateExporter interface {
	WriteEnvironments(Environments) error
	WriteClusters(Clusters) error
	WriteReleases(Releases) error
}

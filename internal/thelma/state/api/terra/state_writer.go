package terra

// StateExporter is an interface for exporting thelma's internal state to some external location
// currently this is sherlock but could also be a file or something else in the future
type StateWriter interface {
	WriteClusters() error
	WriteReleases(Releases) error
}

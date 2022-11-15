package terra

// Named is a top-level interface pulled into Release and Destination types. Having a separate interface representing
// this method means we can iterate over Release and Destination types the same.
type Named interface {
	Name() string
}

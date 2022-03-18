package terra

type Cluster interface {
	Address() string
	Destination
}

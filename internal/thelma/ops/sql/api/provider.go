package api

// Provider cloud provider/infra platform where database is hosted (TODO should probably be moved to state api)
type Provider int64

const (
	Kubernetes Provider = iota
	Google
	Azure
)

func (p Provider) String() string {
	switch p {
	case Kubernetes:
		return "Kubernetes"
	case Google:
		return "Google"
	case Azure:
		return "Azure"
	default:
		return "unknown"
	}
}

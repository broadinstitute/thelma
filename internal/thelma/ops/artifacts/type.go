package artifacts

import "fmt"

type Type int64

const (
	ContainerLog Type = iota
)

func (t Type) pathPrefix() string {
	switch t {
	case ContainerLog:
		return "container-logs"
	default:
		panic(fmt.Sprintf("Unknown artifact type: %v", t))
	}
}

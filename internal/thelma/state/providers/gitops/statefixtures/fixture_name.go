package statefixtures

import "fmt"

// FixtureName is an enum type for different fixtures in the fixtures/ directory.
type FixtureName int

const (
	Default FixtureName = iota
)

func (f FixtureName) String() string {
	switch f {
	case Default:
		return "default"
	}
	panic(fmt.Errorf("unknown fixture: %d", f))
}

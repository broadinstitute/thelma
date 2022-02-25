package logid

import (
	"fmt"
	"math/rand"
)

const maxId = 1<<24 - 1
const format = "%06x"

// NewId return a short, random-enough-for-our-purposes id for identifying events in logs.
// 6 characters long, hexadecimal [0-9a-f]
func NewId() string {
	return fmt.Sprintf(format, rand.Intn(maxId))
}

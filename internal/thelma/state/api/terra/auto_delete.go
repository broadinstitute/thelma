package terra

import "time"

// AutoDelete automatic deletion settings for dynamic environments
type AutoDelete interface {
	// Enabled if true, this environment should be automatically deleted
	Enabled() bool
	// After point in time after which the environment should be automatically deleted
	After() time.Time
}

package terra

import (
	"time"
)

// CreateOptions options for creating a new dynamic environment
type CreateOptions struct {
	// Name to assign to the environment
	Name string
	// GenerateName if true, generate a name instead of using user-supplied name
	GenerateName bool
	// NamePrefix optional prefix to use when generating name
	NamePrefix string
	// AutoDelete optional - if enabled, schedule the BEE for automatic deletion after the given point in time
	AutoDelete struct {
		Enabled bool
		After   time.Time
	}
	// Owner optional - owner to assign to the environment
	Owner string
}

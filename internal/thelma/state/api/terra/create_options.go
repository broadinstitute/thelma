package terra

import (
	"time"
)

// CreateOptions options for creating a new dynamic environment
type CreateOptions struct {
	// Name to assign to the environment
	Name string
	// AutoDelete optional - if enabled, schedule the BEE for automatic deletion after the given point in time
	AutoDelete struct {
		Enabled bool
		After   time.Time
	}
	// Owner optional - owner to assign to the environment
	Owner string

	// StopSchedule an optional daily time to stop the BEE
	StopSchedule struct {
		Enabled       bool
		RepeatingTime time.Time
	}

	// Start Schedule an optional daily time to start the BEE
	StartSchedule struct {
		Enabled       bool
		RepeatingTime time.Time
		Weekends      bool
	}
}

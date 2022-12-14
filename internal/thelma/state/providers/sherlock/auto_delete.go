package sherlock

import "time"

type autoDelete struct {
	enabled bool
	after   time.Time
}

func (a autoDelete) Enabled() bool {
	return a.enabled
}

func (a autoDelete) After() time.Time {
	return a.after
}

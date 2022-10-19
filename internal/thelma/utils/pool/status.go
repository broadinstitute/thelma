package pool

import "github.com/rs/zerolog"

// Status is for representing job status in log messages
type Status struct {
	// Message a short message summarizing job status
	Message string
	// Context contextual field to add to status summary in log messages
	Context map[string]interface{}
}

// Dict returns a representation of this Status as a zerolog Dict for inclusion in log messages
func (s Status) Dict() *zerolog.Event {
	dict := zerolog.Dict()
	dict.Str("message", s.Message)
	for k, v := range s.Context {
		dict.Interface(k, v)
	}
	return dict
}

// StatusReporter is an interface for reporting job status updates
type StatusReporter interface {
	// Update report a job status update
	Update(status Status)
	// channel that worker pool should use to read incoming status updates
	channel() <-chan Status
}

func NewStatusReporter() StatusReporter {
	return statusReporter{
		ch: make(chan Status),
	}
}

type statusReporter struct {
	ch chan Status
}

func (s statusReporter) Update(status Status) {
	s.ch <- status
}

func (s statusReporter) channel() <-chan Status {
	return s.ch
}

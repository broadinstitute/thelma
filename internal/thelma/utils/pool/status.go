package pool

import (
	"github.com/rs/zerolog"
	"sync"
)

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

// StatusReporter is an interface for reporting job status updates in logs. Note that its use is _optional_ -- a job's Run()
// function is free to ignore it.
type StatusReporter interface {
	// Update report a job status update
	Update(status Status)
}

func newStatusReporter() *statusReporter {
	return &statusReporter{
		status:     nil,
		updateCh:   make(chan Status),
		killSwitch: make(chan struct{}),
		mutex:      sync.Mutex{},
	}
}

type statusReporter struct {
	updateCh   chan Status
	status     *Status
	killSwitch chan struct{}
	mutex      sync.Mutex
}

func (s *statusReporter) Update(status Status) {
	s.updateCh <- status
}

func (s *statusReporter) start() {
	go func() {
		for {
			select {
			case status := <-s.updateCh:
				s.mutex.Lock()
				s.status = &status
				s.mutex.Unlock()
			case <-s.killSwitch:
				return
			}
		}
	}()
}

func (s *statusReporter) stop() {
	s.killSwitch <- struct{}{}
}

func (s *statusReporter) getStatus() *Status {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.status
}

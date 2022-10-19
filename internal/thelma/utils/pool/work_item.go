package pool

import (
	"fmt"
	"sync"
	"time"
)

func newWorkItem(job Job, id int) *workItem {
	// set a default name "job-<id>" if one is not supplied by user
	description := job.Description
	if description == "" {
		description = fmt.Sprintf("job-%d", id)
	}

	// make a status reader if the job has a status reporter
	var statusReader *statusReader
	if job.StatusReporter != nil {
		statusReader = newStatusReader(job.StatusReporter)
	}

	return &workItem{
		id:           id,
		name:         description,
		job:          job,
		phase:        Queued,
		statusReader: statusReader,
	}
}

// workItem internal wrapper for Job that includes id & other metadata
type workItem struct {
	job          Job
	name         string
	id           int
	statusReader *statusReader
	phase        Phase
	startTime    time.Time
	endTime      time.Time
	err          error
}

func (w *workItem) execute() {
	w.recordStart()
	err := w.job.Run()
	w.recordStop(err)
}

func (w *workItem) status() *Status {
	if w.statusReader == nil {
		return nil
	}
	return w.statusReader.getStatus()
}

func (w *workItem) recordStart() {
	w.startTime = time.Now()
	w.phase = Running

	if w.statusReader != nil {
		w.statusReader.start()
	}
}

func (w *workItem) recordStop(err error) {
	w.err = err
	w.endTime = time.Now()
	if w.err == nil {
		w.phase = Success
	} else {
		w.phase = Error
	}

	if w.statusReader != nil {
		w.statusReader.stop()
	}
}

// return time the item has been running or total time spent processing, if complete
func (w *workItem) duration() time.Duration {
	if w.phase == Queued {
		return 0
	}
	if w.phase == Running {
		return time.Since(w.startTime)
	}
	return w.endTime.Sub(w.startTime)
}

func newStatusReader(s StatusReporter) *statusReader {
	return &statusReader{
		status:         nil,
		statusUpdateCh: s.channel(),
		killSwitch:     make(chan struct{}),
		mutex:          sync.Mutex{},
	}
}

// statusReader saves the most-recently-reported status for safe reading by other goroutines
type statusReader struct {
	status         *Status
	statusUpdateCh <-chan Status
	killSwitch     chan struct{}
	mutex          sync.Mutex
}

func (s *statusReader) start() {
	go func() {
		for {
			select {
			case status := <-s.statusUpdateCh:
				s.mutex.Lock()
				s.status = &status
				s.mutex.Unlock()
			case <-s.killSwitch:
				return
			}
		}
	}()
}

func (s *statusReader) stop() {
	s.killSwitch <- struct{}{}
}

func (s *statusReader) getStatus() *Status {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.status
}

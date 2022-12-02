package pool

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics"
	"github.com/broadinstitute/thelma/internal/thelma/app/metrics/labels"
	"sync"
	"time"
)

func newWorkItem(job Job, id int, metrics MetricsOptions) workItem {
	// set a default name "job-<id>" if one is not supplied by user
	description := job.Name
	if description == "" {
		description = fmt.Sprintf("job-%d", id)
	}

	return &workItemImpl{
		id:             id,
		name:           description,
		job:            job,
		phase:          Queued,
		statusReporter: newStatusReporter(),
		metrics:        metrics,
	}
}

type workItem interface {
	getName() string
	getId() int
	getPhase() Phase
	getErr() error
	hasErr() bool
	execute()
	status() *Status
	duration() time.Duration
}

// workItemImpl internal wrapper for Job that includes id & other metadata
type workItemImpl struct {
	job            Job
	name           string
	id             int
	statusReporter *statusReporter
	metrics        MetricsOptions
	phase          Phase
	startTime      time.Time
	endTime        time.Time
	err            error
	mutex          sync.Mutex
}

func (w *workItemImpl) getName() string {
	return w.name
}

func (w *workItemImpl) getId() int {
	return w.id
}

func (w *workItemImpl) getPhase() Phase {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.phase
}

func (w *workItemImpl) getErr() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.err
}

func (w *workItemImpl) hasErr() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.err != nil
}

func (w *workItemImpl) execute() {
	w.recordStart()
	err := w.job.Run(w.statusReporter)
	w.recordStop(err)
}

func (w *workItemImpl) status() *Status {
	return w.statusReporter.getStatus()
}

func (w *workItemImpl) recordStart() {
	w.mutex.Lock()
	w.startTime = time.Now()
	w.phase = Running
	w.mutex.Unlock()

	w.statusReporter.start()
}

func (w *workItemImpl) recordStop(err error) {
	w.mutex.Lock()
	w.err = err
	w.endTime = time.Now()
	if w.err == nil {
		w.phase = Success
	} else {
		w.phase = Error
	}
	w.mutex.Unlock()

	w.recordJobMetrics()
	w.statusReporter.stop()
}

func (w *workItemImpl) recordJobMetrics() {
	if !w.metrics.Enabled {
		return
	}

	itemName := w.getName()
	metricName := "pool_" + w.metrics.PoolName

	_labels := map[string]string{
		"pool": w.metrics.PoolName,
		"job":  itemName,
	}

	_labels = labels.Merge(_labels, w.job.Labels)

	opts := metrics.Options{
		Name:   metricName,
		Labels: _labels,
	}

	metrics.TaskCompletion(opts, w.duration(), w.getErr())
}

// return time the item has been running or total time spent processing, if complete
func (w *workItemImpl) duration() time.Duration {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.phase == Queued {
		return 0
	}
	if w.phase == Running {
		return time.Since(w.startTime)
	}
	return w.endTime.Sub(w.startTime)
}

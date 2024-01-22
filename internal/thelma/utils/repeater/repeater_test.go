package repeater

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNew_panics(t *testing.T) {
	assert.Panics(t, func() {
		New(nil)
	})
}

func TestRepeater_startIntervalStop(t *testing.T) {
	var executions int
	var executionsMutex sync.Mutex
	run := func() {
		executionsMutex.Lock()
		executions++
		executionsMutex.Unlock()
	}
	r := New(run, func(o *Options) {
		o.Interval = time.Second / 3
	})
	r.Start()
	time.Sleep(time.Second / 2)
	r.Stop()
	assert.Equal(t, 3, executions)
}

func TestRepeater_startStop(t *testing.T) {
	var executions int
	var executionsMutex sync.Mutex
	run := func() {
		executionsMutex.Lock()
		executions++
		executionsMutex.Unlock()
	}
	r := New(run, func(o *Options) {
		o.Interval = time.Second / 3
	})
	r.Start()
	r.Stop()
	assert.Equal(t, 2, executions)
}

func TestRepeater_start(t *testing.T) {
	var executions int
	var executionsMutex sync.Mutex
	run := func() {
		executionsMutex.Lock()
		executions++
		executionsMutex.Unlock()
	}
	r := New(run, func(o *Options) {
		o.Interval = time.Second / 3
		o.StopRun = false
	})
	r.Start()
	r.Stop()
	assert.Equal(t, 1, executions)
}

func TestRepeater_stop(t *testing.T) {
	var executions int
	var executionsMutex sync.Mutex
	run := func() {
		executionsMutex.Lock()
		executions++
		executionsMutex.Unlock()
	}
	r := New(run, func(o *Options) {
		o.Interval = time.Second / 3
		o.StartRun = false
	})
	r.Start()
	r.Stop()
	assert.Equal(t, 1, executions)
}

func TestRepeater_interval(t *testing.T) {
	var executions int
	var executionsMutex sync.Mutex
	run := func() {
		executionsMutex.Lock()
		executions++
		executionsMutex.Unlock()
	}
	r := New(run, func(o *Options) {
		o.Interval = time.Second / 3
		o.StartRun = false
		o.StopRun = false
	})
	r.Start()
	time.Sleep(time.Second / 2)
	r.Stop()
	assert.Equal(t, 1, executions)
}

package repeater

import (
	"github.com/pkg/errors"
	"time"
)

// Repeater runs an arbitrary side effect at a regular interval.
type Repeater interface {
	// Start begins running the side effect on the configured interval.
	Start()
	// Stop ceases running the side effect.
	Stop()
}

type Options struct {
	// Enabled controls if the side effect should ever be run. If false,
	// Repeater.Start and Repeater.Stop will do nothing.
	Enabled bool
	// Interval controls how often the side effect should be run.
	Interval time.Duration
	// StartRun controls if the side effect should be run immediately when
	// Repeater.Start is called.
	StartRun bool
	// StopRun controls if the side effect should be run immediately when
	// Repeater.Stop is called.
	StopRun bool
}

type Option func(*Options)

// New creates a new Repeater for the given side effect. The interval can be
// configured with the other options; it defaults to 30 seconds.
func New(run func(), options ...Option) Repeater {
	if run == nil {
		panic(errors.Errorf("repeater run function must not be nil"))
	}

	opts := Options{
		Enabled:  true,
		Interval: 30 * time.Second,
		StartRun: true,
		StopRun:  true,
	}

	for _, option := range options {
		option(&opts)
	}

	return &repeater{
		run:        run,
		options:    opts,
		killSwitch: make(chan struct{}),
	}
}

type repeater struct {
	run        func()
	options    Options
	killSwitch chan struct{}
}

func (r *repeater) Start() {
	if r.options.Enabled {
		if r.options.StartRun {
			r.run()
		}

		go func() {
			for {
				select {
				case <-time.After(r.options.Interval):
					r.run()
				case <-r.killSwitch:
					return
				}
			}
		}()
	}
}

func (r *repeater) Stop() {
	if r.options.Enabled {
		r.killSwitch <- struct{}{}

		if r.options.StopRun {
			r.run()
		}
	}
}

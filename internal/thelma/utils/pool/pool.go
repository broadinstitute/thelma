// Package pool contains a generic implementation of the worker pool pattern for concurrent processing
package pool

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Option function for configuring Options
type Option func(*Options)

type Options struct {
	// NumWorkers number of jobs to execute concurrently
	NumWorkers int
	// StopProcessingOnError whether to stop processing work items in the event a job returns an error
	StopProcessingOnError bool
	// Summarizer options for printing periodict processing summaries to the log
	Summarizer SummarizerOptions
}

// Job a unit of work that can be executed by a Pool
type Job struct {
	// Name short text description for this job to use in log messages
	Name string
	// Run function that performs work
	Run func(StatusReporter) error
}

// Pool implements the worker pool pattern for concurrent processing
type Pool interface {
	// Execute starts execution of the pool, returning an error that aggregates errors from all jobs (if any were encountered)
	Execute() error
	// NumWorkers returns the number of workers in the pool
	NumWorkers() int
}

func New(jobs []Job, options ...Option) Pool {
	opts := Options{
		NumWorkers:            runtime.NumCPU(),
		StopProcessingOnError: true,
		Summarizer: SummarizerOptions{
			Enabled:         true,
			Interval:        30 * time.Second,
			LogLevel:        zerolog.InfoLevel,
			WorkDescription: "items processed",
			Footer:          "",
			MaxLineItems:    50,
		},
	}

	for _, option := range options {
		option(&opts)
	}

	var items []workItem
	for i, job := range jobs {
		items = append(items, newWorkItem(job, i))
	}

	cancelCtx, cancelFn := context.WithCancel(context.Background())

	return &pool{
		items:      items,
		options:    opts,
		waitGroup:  sync.WaitGroup{},
		queue:      make(chan workItem, len(items)),
		cancelCtx:  cancelCtx,
		cancelFn:   cancelFn,
		summarizer: newSummarizer(items, opts.Summarizer),
	}
}

type pool struct {
	options    Options
	items      []workItem
	waitGroup  sync.WaitGroup
	queue      chan workItem
	cancelCtx  context.Context
	cancelFn   context.CancelFunc
	summarizer *summarizer
}

func (p *pool) Execute() error {
	log.Debug().Msgf("executing %d job(s) with %d worker(s)", len(p.items), p.NumWorkers())

	p.addJobsToQueue()
	p.summarizer.start()

	for i := 0; i < p.NumWorkers(); i++ {
		p.spawnWorker(i)
	}

	// wait for execution to finish
	p.waitGroup.Wait()
	p.summarizer.stop()

	return p.aggregateErrors()
}

func (p *pool) NumWorkers() int {
	if len(p.items) < p.options.NumWorkers {
		// don't make more workers than we have items to process
		return len(p.items)
	}
	return p.options.NumWorkers
}

func (p *pool) addJobsToQueue() {
	for _, _item := range p.items {
		p.queue <- _item
	}
	close(p.queue)
}

func (p *pool) spawnWorker(id int) {
	p.waitGroup.Add(1)

	logger := log.With().Str("wid", fmt.Sprintf("worker-%d", id)).Logger()

	go func() {
		defer p.waitGroup.Done()

		for {
			select {
			case <-p.cancelCtx.Done():
				logger.Debug().Msg("execution cancelled, returning")
				return

			case item, ok := <-p.queue:
				if !ok {
					logger.Debug().Msg("queue empty, returning")
					return
				}

				itemLogger := logger.With().Str("job", item.getName()).Int("id", item.getId()).Logger()
				itemLogger.Debug().Msg("starting job")
				item.execute()
				itemLogger.Debug().Dur("duration", item.duration()).Str("result", item.getPhase().String()).Err(item.getErr()).Msgf("finished job")

				if item.hasErr() {
					if p.options.StopProcessingOnError {
						logger.Debug().Msgf("cancelling pool")
						p.cancelFn()
						return
					}
				}
			}
		}
	}()
}

// aggregateErrors aggregates all errors into a single mega-error
func (p *pool) aggregateErrors() error {
	var count int
	var sb strings.Builder

	for _, item := range p.items {
		if item.hasErr() {
			count++
			sb.WriteString(fmt.Sprintf("%s: %v\n", item.getName(), item.getErr()))
		}
	}

	if count > 0 {
		return fmt.Errorf("%d execution errors:\n%s", count, sb.String())
	}

	return nil
}

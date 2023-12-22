// Package pool contains a generic implementation of the worker pool pattern for concurrent processing
package pool

import (
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/repeater"
	"github.com/pkg/errors"
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
	// LogSummarizer options for printing periodic processing summaries to the log
	LogSummarizer LogSummarizerOptions
	// Metrics options for recording metrics
	Metrics MetricsOptions
}

type MetricsOptions struct {
	// Enabled if true, record metrics
	Enabled bool
	// PoolName optional name prefix for job metrics
	PoolName string
}

// Job a unit of work that can be executed by a Pool
type Job struct {
	// Name is a short text description for this job to use in log messages
	Name string
	// Run function that performs work
	Run func(StatusReporter) error
	// Labels optional set of metric labels to add to job metrics
	Labels map[string]string
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
		LogSummarizer: LogSummarizerOptions{
			Enabled:         true,
			Interval:        30 * time.Second,
			LogLevel:        zerolog.InfoLevel,
			WorkDescription: "items processed",
			Footer:          "",
			MaxLineItems:    50,
		},
		Metrics: MetricsOptions{
			Enabled:  false,
			PoolName: "unknown",
		},
	}

	for _, option := range options {
		option(&opts)
	}

	var items []workItem
	for i, job := range jobs {
		items = append(items, newWorkItem(job, i, opts.Metrics))
	}

	cancelCtx, cancelFn := context.WithCancel(context.Background())

	return &pool{
		items:                  items,
		options:                opts,
		waitGroup:              sync.WaitGroup{},
		queue:                  make(chan workItem, len(items)),
		cancelCtx:              cancelCtx,
		cancelFn:               cancelFn,
		logSummarizer:          newLogSummarizer(items, opts.LogSummarizer),
	}
}

type pool struct {
	options                Options
	items                  []workItem
	waitGroup              sync.WaitGroup
	queue                  chan workItem
	cancelCtx              context.Context
	cancelFn               context.CancelFunc
	logSummarizer          repeater.Repeater
}

func (p *pool) Execute() error {
	log.Debug().Msgf("executing %d job(s) with %d worker(s)", len(p.items), p.NumWorkers())

	p.addJobsToQueue()
	p.logSummarizer.Start()

	for i := 0; i < p.NumWorkers(); i++ {
		p.spawnWorker(i)
	}

	// wait for execution to finish
	p.waitGroup.Wait()
	p.logSummarizer.Stop()

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
				logger.Trace().Msg("execution cancelled, returning")
				return

			case item, ok := <-p.queue:
				if !ok {
					logger.Trace().Msg("queue empty, returning")
					return
				}

				itemLogger := logger.With().Str("job", item.getName()).Int("id", item.getId()).Logger()
				itemLogger.Trace().Msg("starting job")
				item.execute()
				itemLogger.Trace().Dur("duration", item.duration()).Str("result", item.getPhase().String()).Err(item.getErr()).Msgf("finished job")

				if item.hasErr() {
					if p.options.StopProcessingOnError {
						logger.Trace().Msgf("error encountered; cancelling pool")
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
		return errors.Errorf("%d execution errors:\n%s", count, sb.String())
	}

	return nil
}

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
}

type Job struct {
	// Description text description for this job to use in log messages
	Description string
	// Run function that performs work
	Run func() error
}

// workItem internal wrapper for Job that includes id & other metadata
type workItem struct {
	job         Job
	description string
	id          int
	err         error
	mutex       sync.Mutex
}

// Pool implements the worker pool pattern for concurrent processing
type Pool interface {
	Execute() error
}

func New(jobs []Job, options ...Option) Pool {
	opts := Options{
		NumWorkers:            runtime.NumCPU(),
		StopProcessingOnError: true,
	}
	for _, option := range options {
		option(&opts)
	}

	var items []*workItem
	for i, job := range jobs {
		// set a default description of job-<id>
		description := job.Description
		if description == "" {
			description = fmt.Sprintf("job-%d", i)
		}
		items = append(items, &workItem{
			id:          i,
			description: description,
			job:         job,
		})
	}

	cancelCtx, cancelFn := context.WithCancel(context.Background())

	return &pool{
		items:     items,
		options:   opts,
		waitGroup: sync.WaitGroup{},
		queue:     make(chan *workItem, len(items)),
		cancelCtx: cancelCtx,
		cancelFn:  cancelFn,
	}
}

type pool struct {
	items     []*workItem
	options   Options
	waitGroup sync.WaitGroup
	queue     chan *workItem
	cancelCtx context.Context
	cancelFn  context.CancelFunc
}

func (p *pool) Execute() error {
	log.Debug().Msgf("executing %d job(s) with %d worker(s)", len(p.items), p.numWorkers())

	p.addJobsToQueue()

	for i := 0; i < p.numWorkers(); i++ {
		p.spawnWorker(i)
	}

	// wait for exuection to finish
	p.waitGroup.Wait()

	return p.aggregateErrors()
}

func (p *pool) numWorkers() int {
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

func (p *pool) executeJob(item *workItem, logger zerolog.Logger) {
	item.mutex.Lock()
	defer item.mutex.Unlock()

	logger = logger.With().Str("item", item.description).Logger()

	logger.Debug().Msgf("starting job execution")

	startTime := time.Now()

	item.err = item.job.Run()
	if item.err != nil {
		logger.Err(item.err).Msgf("error executing job")
	}

	logger.Debug().Msgf("job execution finished in %s", time.Since(startTime))
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

				p.executeJob(item, logger)

				if item.err != nil {
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
		if item.err != nil {
			count++
			sb.WriteString(fmt.Sprintf("%s: %v\n", item.description, item.err))
		}
	}

	if count > 0 {
		return fmt.Errorf("%d execution errors:\n%s", count, sb.String())
	}

	return nil
}

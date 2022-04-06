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
	// StopProcessingOnError whether to stop processing work items in the event a job returns an error or times out
	StopProcessingOnError bool
	// Timeout optional timeout for pool's Execute() command
	Timeout time.Duration
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

	if err := p.waitForExecutionToFinish(); err != nil {
		return err
	}

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
						logger.Debug().Msgf("cancelling execution")
						p.cancelFn()
						return
					}
				}
			}
		}
	}()
}

func (p *pool) waitForExecutionToFinish() error {
	if p.options.Timeout <= 0 {
		// no timeout set, so just wait for execution to finish
		p.waitGroup.Wait()
		return nil
	}

	// Wait for workers to finish in a separate goroutine so that we can implement
	// a timeout
	waitCh := make(chan struct{})
	p.launchWorkerWaiter(waitCh)

	// Block until the wait group is done or we time out.
	logger := log.With().Str("wid", "main").Logger()

	select {
	case <-waitCh:
		logger.Debug().Msgf("execution finished")
	case <-time.After(p.options.Timeout):
		err := fmt.Errorf("worker pool execution timed out after %s", p.options.Timeout)
		logger.Error().Err(err)
		logger.Debug().Msgf("cancelling pool")
		p.cancelFn()
		return err
	}

	return nil
}

func (p *pool) launchWorkerWaiter(waitCh chan struct{}) {
	go func() {
		logger := log.With().Str("wid", "wait").Logger()
		logger.Debug().Msg("waiting for workers to finish")
		p.waitGroup.Wait()
		logger.Debug().Msgf("workers finished")
		close(waitCh)
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

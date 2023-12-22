package pool

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/testutils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixMilli())
}

func Test_EmptyWorkloadSucceeds(t *testing.T) {
	require.NoError(t, New([]Job{}).Execute())
}

func Test_SingleJobWithCustomStatusReporter(t *testing.T) {
	job := Job{
		Name: "sam",
		Run: func(s StatusReporter) error {
			for i := 0; i < 10; i++ {
				log.Debug().Msgf("sending status: %d", i)
				s.Update(Status{
					Message: fmt.Sprintf("count: %d", i),
					Context: map[string]interface{}{
						"i": i,
					},
				})
				time.Sleep(10 * time.Millisecond)
			}
			log.Debug().Msg("sending done")
			s.Update(Status{
				Message: "DONE",
			})
			log.Debug().Msgf("return")
			return nil
		},
	}

	require.NoError(t, New([]Job{job}, func(options *Options) {
		options.LogSummarizer.Enabled = true
		options.LogSummarizer.Interval = 5 * time.Millisecond
	}).Execute())
}

func Test_SingleItemSucceeds(t *testing.T) {
	job := successfulJob("job-1")
	require.NoError(t, New(asJobs(job)).Execute())
	assert.Equal(t, 1, job.getCallCount())
}

func Test_MultipleItemsSucceed(t *testing.T) {
	jobs := []*testJob{
		successfulJob("job-1"),
		successfulJob("job-2"),
		successfulJob("job-3"),
	}

	p := New(asJobs(jobs...))
	require.NoError(t, p.Execute())

	for _, j := range jobs {
		assert.Equal(t, 1, j.getCallCount(), "job %s should have been called exactly once", j.description)
	}
}

func Test_SingleItemFailureStopsProcessing(t *testing.T) {
	j1 := successfulJob("job-1")
	j2 := failingJob("job-2")
	j3 := successfulJob("job-3")

	p := New(asJobs(j1, j2, j3), func(options *Options) {
		options.NumWorkers = 1
	})
	err := p.Execute()
	assert.Error(t, err)
	assert.Equal(t, "1 execution errors:\njob-2: whoopsies (job-2)\n", err.Error())

	assert.Equal(t, 1, j1.getCallCount(), "job 1 should have been called exactly once")
	assert.Equal(t, 1, j2.getCallCount(), "job 2 should have been called exactly once")
	assert.Equal(t, 0, j3.getCallCount(), "job 3 should not have been called")
}

func Test_SingleItemFailureContinuesProcessingIfStopIsFalse(t *testing.T) {
	j1 := successfulJob("job-1")
	j2 := failingJob("job-2")
	j3 := successfulJob("job-3")

	p := New(asJobs(j1, j2, j3), func(options *Options) {
		options.StopProcessingOnError = false
		options.NumWorkers = 1
	})
	err := p.Execute()
	assert.Error(t, err)
	assert.Equal(t, "1 execution errors:\njob-2: whoopsies (job-2)\n", err.Error())

	assert.Equal(t, 1, j1.getCallCount(), "job 1 should have been called exactly once")
	assert.Equal(t, 1, j2.getCallCount(), "job 2 should have been called exactly once")
	assert.Equal(t, 1, j3.getCallCount(), "job 3 should have been called exactly once")
}

func Test_LargeBatchCompletes(t *testing.T) {
	var jobs []*testJob
	for i := 0; i < 1000; i++ {
		jobs = append(jobs, sleepingJob(fmt.Sprintf("job-%d", i), randomIntervalUnder100ms()))
	}

	p := New(asJobs(jobs...), func(options *Options) {
		options.NumWorkers = 10
	})
	err := p.Execute()
	assert.NoError(t, err)

	var numCalled, numFailed int
	for _, job := range jobs {
		if job.err != nil {
			numFailed++
		}
		if job.getCallCount() > 0 {
			numCalled++
		}
	}

	assert.Equal(t, 0, numFailed)
	assert.Equal(t, numCalled, 1000)
}

func Test_LargeBatchStopsExecutingOnFailure(t *testing.T) {
	var jobs []*testJob
	for i := 0; i < 1000; i++ {
		if i == 230 {
			jobs = append(jobs, failingJob(fmt.Sprintf("job-%d", i)))
		} else {
			jobs = append(jobs, sleepingJob(fmt.Sprintf("job-%d", i), randomIntervalUnder100ms()))
		}
	}

	p := New(asJobs(jobs...), func(options *Options) {
		options.NumWorkers = 10
	})
	err := p.Execute()
	assert.Error(t, err)
	assert.Equal(t, "1 execution errors:\njob-230: whoopsies (job-230)\n", err.Error())

	var numCalled, numFailed int
	for _, job := range jobs {
		if job.err != nil {
			numFailed++
		}
		if job.getCallCount() > 0 {
			numCalled++
		}
	}

	assert.Equal(t, 1, numFailed)
	assert.Greater(t, numCalled, 200)
	assert.Less(t, numCalled, 300)
}

type testJob struct {
	description string
	err         error
	callCount   int
	sleep       time.Duration
	mutex       sync.Mutex
}

func (j *testJob) job() Job {
	return Job{
		Name: j.description,
		Run:  j.run,
	}
}

func (j *testJob) run(s StatusReporter) error {
	j.mutex.Lock()
	j.callCount++
	j.mutex.Unlock()

	if j.sleep == 0 {
		return j.err
	}

	// take the total time for our fake test job and split it into 5 random intervals.
	// then send a status update after each interval.
	// the goal here is to thoroughly exercise the status updater to catch concurrency issues
	intervals := testutils.SliceIntoRandomIntervals(j.sleep, rand.Intn(5))
	for i, interval := range intervals {
		msg := "even"
		if i%2 == 0 {
			msg = "odd"
		}
		s.Update(Status{
			Message: msg,
			Context: map[string]interface{}{"i": i, "n": len(intervals)},
		})
		time.Sleep(interval)
	}
	return j.err
}

func (j *testJob) getCallCount() int {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.callCount
}

func successfulJob(desc string) *testJob {
	return &testJob{
		description: desc,
	}
}

func failingJob(desc string) *testJob {
	return &testJob{
		description: desc,
		err:         errors.Errorf("whoopsies (%s)", desc),
	}
}

func sleepingJob(desc string, sleepTime time.Duration) *testJob {
	return &testJob{
		description: desc,
		sleep:       sleepTime,
	}
}

func asJobs(jobs ...*testJob) []Job {
	result := make([]Job, len(jobs))
	for i, j := range jobs {
		result[i] = j.job()
	}
	return result
}

func randomIntervalUnder100ms() time.Duration {
	multiplier := rand.Int63() % 100
	return time.Duration(multiplier) * time.Millisecond
}

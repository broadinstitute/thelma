package pool

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func Test_ChartReleaseSummarizer(t *testing.T) {
	// Whether the summarizer function errors or not, the behavior should be the same
	for testName, summarizerShouldError := range map[string]bool{
		"summarizer works properly": false,
		"summarizer errors":         true,
	} {
		t.Run(testName, func(t *testing.T) {
			baseInterval := 10 * time.Millisecond
			leonardoDev := Job{
				Name:             "leonardo",
				ChartReleaseName: "leonardo-dev",
				Run: func(s StatusReporter) error {
					s.Update(Status{
						Message: "waiting",
					})
					time.Sleep(8 * baseInterval)
					s.Update(Status{
						Message: "synced",
					})
					return nil
				},
			}
			samDev := Job{
				Name:             "sam",
				ChartReleaseName: "sam-dev",
				Run: func(s StatusReporter) error {
					s.Update(Status{
						Message: "waiting",
					})
					time.Sleep(8 * baseInterval)
					s.Update(Status{
						Message: "progressing",
					})
					time.Sleep(8 * baseInterval)
					s.Update(Status{
						Message: "synced",
					})
					return nil
				},
			}

			var mutex sync.Mutex
			summaries := make([]map[string]string, 0)
			_pool := New([]Job{leonardoDev, samDev}, func(options *Options) {
				options.NumWorkers = 2
				options.LogSummarizer.Enabled = false
				options.ChartReleaseSummarizer.Enabled = true
				options.ChartReleaseSummarizer.Interval = 5 * baseInterval
				options.ChartReleaseSummarizer.Do = func(statuses map[string]string) error {
					mutex.Lock()
					summaries = append(summaries, statuses)
					mutex.Unlock()
					if summarizerShouldError {
						return errors.Errorf("some error from the summarizer do function")
					} else {
						return nil
					}
				}
			})
			err := _pool.Execute()
			assert.NoError(t, err)
			assert.Equal(t, []map[string]string{
				{"leonardo-dev": "queued", "sam-dev": "queued"},
				{"leonardo-dev": "running: waiting", "sam-dev": "running: waiting"},
				{"leonardo-dev": "success: synced", "sam-dev": "running: progressing"},
				{"leonardo-dev": "success: synced", "sam-dev": "running: progressing"},
				{"leonardo-dev": "success: synced", "sam-dev": "success: synced"},
			}, summaries)
		})
	}
}

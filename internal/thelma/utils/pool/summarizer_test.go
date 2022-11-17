package pool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
	"time"
)

func Test_Summarizer(t *testing.T) {
	file := path.Join(t.TempDir(), "log")
	writer, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err)
	logger := zerolog.New(writer)

	carrot := Job{
		Name: "carrot",
		Run: func(s StatusReporter) error {
			s.Update(Status{
				Message: "rabbit gnawing",
				Context: map[string]interface{}{
					"name": "peter",
				},
			})
			time.Sleep(8 * time.Millisecond)
			s.Update(Status{
				Message: "rabbit full",
				Context: map[string]interface{}{
					"name": "peter",
				},
			})

			return nil
		},
	}

	celery := Job{
		Name: "celery",
		Run: func(s StatusReporter) error {
			s.Update(Status{
				Message: "1% complete",
			})
			time.Sleep(2 * time.Millisecond)
			s.Update(Status{
				Message: "12% complete",
			})
			time.Sleep(5 * time.Millisecond)
			s.Update(Status{
				Message: "67% complete",
			})
			time.Sleep(5 * time.Millisecond)
			s.Update(Status{
				Message: "100% complete",
			})
			return nil
		},
	}

	onion := Job{
		Name: "onion",
		Run: func(s StatusReporter) error {
			s.Update(Status{
				Message: "onion pending",
			})
			time.Sleep(5 * time.Millisecond)
			return fmt.Errorf("whoopsies")
		},
	}

	_pool := New([]Job{carrot, celery, onion}, func(options *Options) {
		options.NumWorkers = 2
		options.Summarizer.Enabled = true
		options.Summarizer.Interval = 5 * time.Millisecond
		options.Summarizer.WorkDescription = "veggies eaten"
		options.Summarizer.Footer = "check https://veggies.broadinstitute.org for updates"
		options.Summarizer.LogLevel = zerolog.WarnLevel
		options.Summarizer.logger = &logger
	})
	err = _pool.Execute()
	assert.ErrorContains(t, err, "onion: whoopsies")

	messages := parseMessages(t, file)
	assert.Equal(t, []map[string]interface{}{
		// update #1 -- before processing starts
		{
			"level":   "warn",
			"queued":  3,
			"message": "0/3 veggies eaten",
		},
		{
			"level":   "warn",
			"message": "carrot: queued",
		},
		{
			"level":   "warn",
			"message": "celery: queued",
		},
		{
			"level":   "warn",
			"message": "onion:  queued",
		},
		{
			"level":   "warn",
			"message": "check https://veggies.broadinstitute.org for updates",
		},

		// update #2 -- during processing
		{
			"level":   "warn",
			"running": 2,
			"queued":  1,
			"message": "0/3 veggies eaten",
		},
		{
			"level":   "warn",
			"message": "carrot: running",
			"status":  "rabbit gnawing",
			"name":    "peter",
		},
		{
			"level":   "warn",
			"message": "celery: running",
			"status":  "12% complete",
		},
		{
			"level":   "warn",
			"message": "onion:  queued",
		},
		{
			"level":   "warn",
			"message": "check https://veggies.broadinstitute.org for updates",
		},

		// update #3 -- during processing
		{
			"level":   "warn",
			"running": 2,
			"success": 1,
			"message": "1/3 veggies eaten",
		},
		{
			"level":   "warn",
			"message": "carrot: success",
			"status":  "rabbit full",
			"name":    "peter",
		},
		{
			"level":   "warn",
			"message": "celery: running",
			"status":  "67% complete",
		},
		{
			"level":   "warn",
			"message": "onion:  running",
			"status":  "onion pending",
		},
		{
			"level":   "warn",
			"message": "check https://veggies.broadinstitute.org for updates",
		},

		// final status update -- after processing
		{
			"level":   "warn",
			"success": 2,
			"error":   1,
			"message": "3/3 veggies eaten",
		},
		{
			"level":   "warn",
			"message": "carrot: success",
			"status":  "rabbit full",
			"name":    "peter",
		},
		{
			"level":   "warn",
			"message": "celery: success",
			"status":  "100% complete",
		},
		{
			"level":   "warn",
			"message": "onion:  error",
			"error":   "whoopsies",
			"status":  "onion pending",
		},
		{
			"level":   "warn",
			"message": "check https://veggies.broadinstitute.org for updates",
		},
	}, messages)
}

func parseMessages(t *testing.T, file string) []map[string]interface{} {
	content, err := os.ReadFile(file)
	require.NoError(t, err)

	fmt.Println(string(content))
	var messages []map[string]interface{}
	for _, line := range bytes.Split(content, []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		msg := make(map[string]interface{})

		require.NoError(t, json.Unmarshal(line, &msg))

		// json will decode numebrs into float64s -- convert values to int before we compare
		for k, v := range msg {
			if asFloat, ok := v.(float64); ok {
				msg[k] = int(asFloat)
			}
		}

		// remove duration field because it's unpredictable so we can't assert on it
		delete(msg, elapsedTimeField)

		messages = append(messages, msg)
	}

	return messages
}

package testutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_SliceIntoRandomIntervals(t *testing.T) {
	for i := 1; i < 100; i++ {
		intervals := SliceIntoRandomIntervals(time.Second, i)
		assert.Equal(t, i, len(intervals))

		var sum time.Duration
		for _, interval := range intervals {
			sum = sum + interval
		}
		assert.Equal(t, time.Second, sum)
	}
}

package testutils

import (
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"sort"
	"strings"
	"time"
)

var alphaNumeric = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Args convenience function to generate tokenized argument list from format string w/ args
//
// Eg. args("-e   %s", "dev") -> []string{"-e", "dev"}
func Args(format string, a ...interface{}) []string {
	formatted := fmt.Sprintf(format, a...)
	return strings.Fields(formatted)
}

// RandString generates a random alphanumeric string (a-z0-9) of length n
func RandString(n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = alphaNumeric[rand.Intn(len(alphaNumeric))]
	}
	return string(result)
}

// SliceIntoRandomIntervals slices a time.Duration into n random intervals
func SliceIntoRandomIntervals(duration time.Duration, n int) []time.Duration {
	if n < 0 {
		panic(errors.Errorf("can't divide duration into %d intervals", n))
	}
	asInt64 := int64(duration)

	var boundaries []int64
	boundaries = append(boundaries, 0)
	for i := 0; i < n-1; i++ {
		boundaries = append(boundaries, rand.Int63n(asInt64))
	}
	boundaries = append(boundaries, asInt64)

	sort.Slice(boundaries, func(i, j int) bool {
		return boundaries[i] < boundaries[j]
	})

	var result []time.Duration
	for i := 1; i < len(boundaries); i++ {
		delta := boundaries[i] - boundaries[i-1]
		result = append(result, time.Duration(delta))
	}

	return result
}

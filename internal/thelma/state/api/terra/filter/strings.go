package filter

import (
	"fmt"
	"strings"
)

// this file constants and helpers for string representations of filters
const andFormat = "and(%s)"
const orFormat = "or(%s)"
const notFormat = "not(%s)"
const anyString = "any()"

func quote(strings []string) []string {
	var result []string
	for _, s := range strings {
		result = append(result, fmt.Sprintf("%q", s))
	}
	return result
}

// join a set of strings that represent a range of possible values in a filter
func join(elements ...string) string {
	return strings.Join(elements, ",")
}

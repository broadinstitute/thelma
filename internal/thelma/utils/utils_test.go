package utils

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func Test_Not(t *testing.T) {
	isEven := func(n int) bool {
		return n%2 == 0
	}

	assert.False(t, Not(isEven)(0))
	assert.True(t, Not(isEven)(1))
	assert.False(t, Not(isEven)(2))
}

func Test_MapValues(t *testing.T) {
	input := map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
	}
	expected := []string{"a", "b", "c"}
	actual := MapValues(input)
	sort.Strings(actual)
	assert.Equal(t, expected, actual)
}

func Test_MapValuesFlattened(t *testing.T) {
	input := map[string][]string{
		"1": {"a", "a", "a"},
		"2": {"b"},
		"3": {"c", "c"},
	}
	expected := []string{"a", "a", "a", "b", "c", "c"}
	actual := MapValuesFlattened(input)
	sort.Strings(actual)
	assert.Equal(t, expected, actual)
}

func Test_MapKeys(t *testing.T) {
	input := map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
	}
	expected := []string{"1", "2", "3"}
	actual := MapKeys(input)
	sort.Strings(actual)
	assert.Equal(t, expected, actual)
}

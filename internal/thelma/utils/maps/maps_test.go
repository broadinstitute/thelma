package maps

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func Test_MapValues(t *testing.T) {
	input := map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
	}
	expected := []string{"a", "b", "c"}
	actual := Values(input)
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
	actual := ValuesFlattened(input)
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
	actual := Keys(input)
	sort.Strings(actual)
	assert.Equal(t, expected, actual)
}

func Test_TransformKeys(t *testing.T) {
	input := map[string]string{
		"  1  ":  "a",
		" 2 ":    "b",
		"   3\n": "c",
	}

	expected := map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
	}

	actual := TransformKeys(input, func(s string) string {
		return strings.TrimSpace(s)
	})

	assert.Equal(t, expected, actual)
}

func Test_TransformValues(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	expected := map[string]string{
		"a": "3",
		"b": "4",
		"c": "5",
	}

	actual := TransformValues(input, func(n int) string {
		return strconv.Itoa(n + 2)
	})

	assert.Equal(t, expected, actual)
}

func Test_Transform(t *testing.T) {
	input := map[string]int{
		"1": 4,
		"2": 5,
		"3": 6,
	}

	expected := map[int]string{
		10: "8",
		20: "10",
		30: "12",
	}

	actual := Transform(input, func(k string, v int) (int, string) {
		n, err := strconv.Atoi(k)
		require.NoError(t, err)
		k2 := n * 10
		v2 := strconv.Itoa(v * 2)
		return k2, v2
	})

	assert.Equal(t, expected, actual)
}

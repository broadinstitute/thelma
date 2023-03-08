package set

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestSet_BasicOperations(t *testing.T) {
	s := NewSet("a", "a", "b")
	assert.False(t, s.Empty())
	assert.Equal(t, 2, s.Size())
	assert.True(t, s.Exists("a"))
	assert.True(t, s.Exists("b"))

	elts := s.Elements()
	sort.Strings(elts)
	assert.Equal(t, []string{"a", "b"}, elts)

	// add 2 new elements
	s.Add("c", "a", "d", "b", "b", "d")
	assert.Equal(t, 4, s.Size())
	assert.True(t, s.Exists("c"))
	assert.True(t, s.Exists("d"))

	elts = s.Elements()
	sort.Strings(elts)
	assert.Equal(t, []string{"a", "b", "c", "d"}, elts)

	// remove elements
	s.Remove("a", "z")
	assert.Equal(t, 3, s.Size())
	assert.False(t, s.Exists("a"))
	assert.True(t, s.Exists("b"))
	assert.True(t, s.Exists("c"))
	assert.True(t, s.Exists("d"))

	elts = s.Elements()
	sort.Strings(elts)
	assert.Equal(t, []string{"b", "c", "d"}, elts)

}

func TestSet_Difference(t *testing.T) {
	testCases := []struct {
		name     string
		set      []string
		diff     []string
		expected []string
	}{
		{
			name:     "empty",
			expected: []string{},
		},
		{
			name:     "one elt",
			set:      []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "one elt in dff",
			set:      []string{"a"},
			diff:     []string{"a"},
			expected: []string{},
		},
		{
			name:     "extra elt",
			set:      []string{"a"},
			diff:     []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "complex",
			set:      []string{"a", "b", "c", "d"},
			diff:     []string{"a", "c", "e", "f"},
			expected: []string{"b", "d"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSet(tc.set...)
			diff := NewSet(tc.diff...)

			result := s.Difference(diff).Elements()
			sort.Strings(result)
			assert.Equal(t, tc.expected, result)
		})
	}
}

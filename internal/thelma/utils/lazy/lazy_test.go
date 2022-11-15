package lazy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Lazy(t *testing.T) {
	counter := 0
	l := NewLazy[int](func() int {
		counter++
		return counter
	})

	assert.Equal(t, 1, l.Get(), "should call initializer")
	assert.Equal(t, 1, l.Get(), "should only call initializer once")
}

func Test_LazyE(t *testing.T) {
	counter := 0
	l := NewLazyE[int](func() (int, error) {
		counter++
		if counter < 3 {
			return counter, fmt.Errorf("counter < 3")
		}
		return counter, nil
	})

	v, err := l.Get()
	require.ErrorContains(t, err, "counter < 3")
	assert.Equal(t, 1, v)

	v, err = l.Get()
	require.ErrorContains(t, err, "counter < 3")
	assert.Equal(t, 2, v)

	v, err = l.Get()
	require.NoError(t, err)
	assert.Equal(t, 3, v)

	v, err = l.Get()
	require.NoError(t, err)
	assert.Equal(t, 3, v, "should return cached value after successful initialization")
}

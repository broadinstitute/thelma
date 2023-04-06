package utils

import (
	"github.com/stretchr/testify/assert"
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

func Test_JoinSelector(t *testing.T) {
	assert.Equal(t, "", JoinSelector(map[string]string{}))
	assert.Equal(t, "a=b,c=d,x=y", JoinSelector(map[string]string{"x": "y", "a": "b", "c": "d"}))
}

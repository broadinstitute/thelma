package lazy

import "sync"

// LazyE is a caching, concurrent-safe lazy initializer, for initializations that might produce an error
type LazyE[T any] interface {
	// Get when first called, will run the initialization function and cache the output. Returns cached output on
	// subsequent runs.
	Get() (T, error)
}

// NewLazyE given an initialization function return a LazyE that wraps it
func NewLazyE[T any](initializer func() (T, error)) LazyE[T] {
	return &lazye[T]{
		initializer: initializer,
	}
}

type lazye[T any] struct {
	initializer func() (T, error)
	value       *T
	mutex       sync.Mutex
}

func (l *lazye[T]) Get() (T, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.value != nil {
		return *l.value, nil
	}

	value, err := l.initializer()
	if err != nil {
		return value, err
	}

	l.value = &value

	return value, nil
}

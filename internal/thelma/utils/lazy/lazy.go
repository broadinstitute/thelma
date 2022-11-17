package lazy

// Lazy is a caching concurrency-safe lazy initializer
type Lazy[T any] interface {
	// Get when first called, will run the initialization function and cache the output. Returns cached output on
	// subsequent runs.
	Get() T
}

// NewLazy given an initialization function return a Lazy that wraps it
func NewLazy[T any](initializer func() T) Lazy[T] {
	return &lazy[T]{
		lazye: NewLazyE(func() (T, error) {
			return initializer(), nil
		}),
	}
}

type lazy[T any] struct {
	lazye LazyE[T]
}

func (l *lazy[T]) Get() T {
	v, _ := l.lazye.Get()
	return v
}

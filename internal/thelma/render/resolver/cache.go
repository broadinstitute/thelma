package resolver

import (
	"strings"
	"sync"
)

// Separator use when joining elements of a ChartRelease into a string cache key
const defaultCacheKeySeparator = " " // URLs can't have whitespace, so this won't show up
// in Helm repo and chart names, or in chart version strings

// syncCache is a concurrent-safe cache for ResolvedCharts. Guarantees a chart won't be resolved multiple times.
//
// The important complexity here is that different resolvers want to resolve charts from different input:
//   - remoteResolver, for instance, wants the cache to operate per ChartRelease.
//   - localResolver, by contrast, wants the cache to operate per on-disk Chart, regardless of what environment it is
//     being released to.
//
// syncCache is generic on the input type so that it can accomplish either case with the same important mutex/locking
// behavior.
type syncCache[R any] interface {
	get(resolvable R) (ResolvedChart, error)
}

// keyMapper determines how syncCache should resolve the resolvable input to a cache key.
// In other words, this determines when syncCache will actually call resolver on the resolvable versus getting a
// cache hit.
type keyMapper[R any] func(resolvable R) string

// resolver is responsible for converting an (uncached) resolvable to a ResolvedChart.
type resolver[R any] func(resolvable R) (ResolvedChart, error)

// chartReleaseKeyMapper is the default keyMapper for a syncCache of ChartRelease (the usual case)
func chartReleaseKeyMapper(chartRelease ChartRelease) string {
	return strings.Join(
		[]string{
			chartRelease.Repo,
			chartRelease.Name,
			chartRelease.Version,
		},
		defaultCacheKeySeparator,
	)
}

type syncCacheImpl[R any] struct {
	globalMutex sync.RWMutex
	keyMapper   keyMapper[R]
	resolver    resolver[R]
	cache       map[string]*entry[R]
}

// Returns a new syncCache instance
func newSyncCache(resolver resolver[ChartRelease]) syncCache[ChartRelease] {
	return newSyncCacheWithMapper(resolver, chartReleaseKeyMapper)
}

func newSyncCacheWithMapper[R any](resolver resolver[R], keyMapper keyMapper[R]) syncCache[R] {
	return &syncCacheImpl[R]{
		globalMutex: sync.RWMutex{},
		keyMapper:   keyMapper,
		resolver:    resolver,
		cache:       make(map[string]*entry[R]),
	}
}

//nolint:unused
func (c *syncCacheImpl[R]) get(resolvable R) (ResolvedChart, error) {
	key := c.keyMapper(resolvable)
	_entry := c.getEntry(key)

	return _entry.getValue(resolvable, c.resolver)
}

//nolint:unused
func (c *syncCacheImpl[R]) getEntry(key string) *entry[R] {
	// get a read lock so we can safely read from the map
	c.globalMutex.RLock()
	_entry, exists := c.cache[key]
	c.globalMutex.RUnlock()

	if exists {
		return _entry
	}

	// entry does not exist in map, so set a write lock and create one
	c.globalMutex.Lock()
	defer c.globalMutex.Unlock()

	_entry = &entry[R]{}
	c.cache[key] = _entry

	return _entry
}

// Represents an entry in the cache
//
//nolint:unused
type entry[R any] struct {
	initialized   bool
	mutex         sync.RWMutex
	resolvedChart ResolvedChart
	err           error
}

// Returns the cached value for a given entry
//
//nolint:unused
func (e *entry[R]) getValue(resolvable R, resolver resolver[R]) (ResolvedChart, error) {
	// Obtain a read lock
	// if we've been initialized already, return the cached values
	e.mutex.RLock()
	if e.initialized {
		defer e.mutex.RUnlock()
		return e.resolvedChart, e.err
	}

	// Otherwise: release the read lock & obtain a write lock instead
	e.mutex.RUnlock()
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// generate the values
	rc, err := resolver(resolvable)

	// save them & mark this entry as initialized
	e.resolvedChart = rc
	e.err = err
	e.initialized = true

	// return result
	return e.resolvedChart, e.err
}

package logs

import (
	"fmt"
	"path"
	"sync"
)

func newLogNameGenerator() *logNameGenerator {
	return &logNameGenerator{cache: make(map[string]int)}
}

type logNameGenerator struct {
	cache map[string]int
	mutex sync.Mutex
}

// generate a unique-to-release container log name
// defaults to "<container-name>.log", but adds incrementing suffix for additional containers w/ same name
func (c *logNameGenerator) generateName(container container) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cacheKey := path.Join("%s:%s", container.release.Name(), container.containerName)
	count := c.cache[cacheKey]

	var name string
	if count == 0 {
		name = fmt.Sprintf("%s.log", container.containerName)
	} else {
		name = fmt.Sprintf("%s-%d.log", container.containerName, count)
	}

	c.cache[cacheKey] = count + 1

	return name
}

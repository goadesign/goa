package codegen

import (
	"sync"
)

type (
	// stringCache provides memoization for expensive string operations.
	// Used to cache results of CamelCase, Goify, and other string transformations
	// that are called repeatedly with the same inputs during code generation.
	stringCache struct {
		mu    sync.RWMutex
		cache map[cacheKey]string
	}

	// cacheKey uniquely identifies a cached string transformation.
	cacheKey struct {
		input      string
		firstUpper bool
		acronym    bool
		operation  string // "camel", "goify", etc.
	}
)

var (
	// Global cache instance used across code generation
	globalStringCache = &stringCache{
		cache: make(map[cacheKey]string),
	}
)

// getCached retrieves a cached result or computes and caches it.
func (c *stringCache) getCached(key cacheKey, compute func() string) string {
	// Fast path: read lock for cache hit
	c.mu.RLock()
	if result, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		return result
	}
	c.mu.RUnlock()

	// Slow path: compute and cache
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check in case another goroutine computed it
	if result, ok := c.cache[key]; ok {
		return result
	}

	result := compute()
	c.cache[key] = result
	return result
}

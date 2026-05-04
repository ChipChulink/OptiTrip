package cache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     interface{}
	timestamp time.Time
	ttl       time.Duration
}

type Cache struct {
	data  map[string]cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
	hits  int64
	misses int64
}

func New(ttlMinutes int) *Cache {
	return &Cache{
		data: make(map[string]cacheEntry),
		ttl:  time.Duration(ttlMinutes) * time.Minute,
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		c.misses++
		return nil, false
	}

	if time.Since(entry.timestamp) > entry.ttl {
		c.misses++
		delete(c.data, key)
		return nil, false
	}

	c.hits++
	return entry.value, true
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		value:     value,
		timestamp: time.Now(),
		ttl:       c.ttl,
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
}

func (c *Cache) Stats() (hits, misses int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

func (c *Cache) InvalidateByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

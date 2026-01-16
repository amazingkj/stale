package cache

import (
	"sync"
	"time"
)

// Item represents a cached item with expiration
type Item[T any] struct {
	Value      T
	Expiration int64
}

// Cache is a simple in-memory cache with TTL support
type Cache[T any] struct {
	mu    sync.RWMutex
	items map[string]Item[T]
	ttl   time.Duration
}

// New creates a new cache with the specified TTL
func New[T any](ttl time.Duration) *Cache[T] {
	c := &Cache[T]{
		items: make(map[string]Item[T]),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves an item from the cache
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		var zero T
		return zero, false
	}

	if time.Now().UnixNano() > item.Expiration {
		var zero T
		return zero, false
	}

	return item.Value, true
}

// Set stores an item in the cache
func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item[T]{
		Value:      value,
		Expiration: time.Now().Add(c.ttl).UnixNano(),
	}
}

// Delete removes an item from the cache
func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item[T])
}

// cleanup periodically removes expired items
func (c *Cache[T]) cleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

package cache

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Entry is a cached response with metadata.
type Entry struct {
	Body       []byte
	Headers    map[string][]string
	StatusCode int
	CreatedAt  time.Time
}

// Stats holds cache statistics for the analytics dashboard.
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Entries    int   `json:"entries"`
	MaxEntries int   `json:"max_entries"`
}

// Cache is an in-memory LRU response cache with TTL expiration.
type Cache struct {
	mu         sync.RWMutex
	maxEntries int
	ttl        time.Duration
	items      map[string]*list.Element
	order      *list.List // LRU order
	hits       int64
	misses     int64
}

type cacheItem struct {
	key   string
	entry *Entry
}

// New creates a new response cache. If maxEntries is 0, cache is disabled.
func New(maxEntries int, ttlSeconds int) *Cache {
	if maxEntries <= 0 {
		return nil
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Cache{
		maxEntries: maxEntries,
		ttl:        ttl,
		items:      make(map[string]*list.Element),
		order:      list.New(),
	}
}

// Hash generates a cache key from a request body using SHA-256.
func Hash(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

// Get retrieves a cached response. Returns nil if not found or expired.
func (c *Cache) Get(key string) *Entry {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		c.misses++
		return nil
	}

	item := elem.Value.(*cacheItem)

	// Check TTL expiration.
	if time.Since(item.entry.CreatedAt) > c.ttl {
		c.order.Remove(elem)
		delete(c.items, key)
		c.misses++
		return nil
	}

	// Move to front (most recently used).
	c.order.MoveToFront(elem)
	c.hits++
	return item.entry
}

// Set stores a response in the cache. Evicts LRU entry if full.
func (c *Cache) Set(key string, entry *Entry) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry.
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*cacheItem).entry = entry
		return
	}

	// Evict LRU if at capacity.
	if c.order.Len() >= c.maxEntries {
		back := c.order.Back()
		if back != nil {
			evicted := c.order.Remove(back).(*cacheItem)
			delete(c.items, evicted.key)
		}
	}

	// Insert new entry.
	item := &cacheItem{key: key, entry: entry}
	elem := c.order.PushFront(item)
	c.items[key] = elem
}

// GetStats returns cache hit/miss stats for the dashboard.
func (c *Cache) GetStats() Stats {
	if c == nil {
		return Stats{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Hits:       c.hits,
		Misses:     c.misses,
		Entries:    c.order.Len(),
		MaxEntries: c.maxEntries,
	}
}

// Package cache is refactoring exercise 06: concurrency + abstraction (hard).
//
// GOAL
//
//	Make this TTL cache concurrent-friendly and testable without breaking its
//	two guarantees, which the tests pin:
//	  1. a value is loaded at most once per key while it is fresh, even when
//	     many goroutines ask for the same missing key at the same instant;
//	  2. a value expires after its TTL.
//
//	Run both:
//	    go test ./...
//	    go test -race ./...
//
// SMELLS TO HUNT
//
//   - one coarse mutex held across the SLOW loader call. It makes the whole
//     cache single-threaded: a slow load for key "a" blocks an unrelated Get
//     of key "b". Reducing the lock scope is the goal — but do it WITHOUT
//     breaking guarantee #1. The naive fix (unlock, load, relock) lets N
//     concurrent callers all run the loader N times; the "load once" test
//     will catch that. You need per-key coordination (think singleflight:
//     one caller loads, the rest wait for that result).
//   - no finer-grained or per-key locking, so there is no path to real
//     concurrency across different keys.
//
// The clock is already injected (the `now` field) so TTL is testable without
// real sleeps — use that as the model for keeping dependencies explicit.
package cache

import (
	"sync"
	"time"
)

// Loader produces the value for a key on a cache miss. It is assumed to be
// expensive.
type Loader func(key string) int

type entry struct {
	value   int
	expires time.Time
}

// Cache is a tiny TTL cache guarding an expensive Loader.
type Cache struct {
	mu   sync.Mutex
	data map[string]entry
	ttl  time.Duration
	load Loader
	now  func() time.Time
}

// New builds a cache. now may be nil, in which case time.Now is used.
func New(ttl time.Duration, load Loader, now func() time.Time) *Cache {
	if now == nil {
		now = time.Now
	}
	return &Cache{
		data: make(map[string]entry),
		ttl:  ttl,
		load: load,
		now:  now,
	}
}

// Get returns the cached value for key, loading it on a miss or after expiry.
func (c *Cache) Get(key string) int {
	// The lock is held across the entire method, including the slow load.
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.data[key]
	if ok && c.now().Before(e.expires) {
		return e.value
	}

	value := c.load(key) // SLOW, and we are holding the one global lock.
	c.data[key] = entry{
		value:   value,
		expires: c.now().Add(c.ttl),
	}
	return value
}

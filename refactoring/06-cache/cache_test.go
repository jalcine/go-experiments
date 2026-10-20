package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetCachesValue(t *testing.T) {
	var calls int32
	c := New(time.Minute, func(key string) int {
		atomic.AddInt32(&calls, 1)
		return 42
	}, func() time.Time { return time.Unix(0, 0) })

	if got := c.Get("a"); got != 42 {
		t.Fatalf("Get = %d, want 42", got)
	}
	if got := c.Get("a"); got != 42 {
		t.Fatalf("Get (cached) = %d, want 42", got)
	}
	if calls != 1 {
		t.Errorf("loader called %d times, want 1 (second Get should hit cache)", calls)
	}
}

func TestGetExpires(t *testing.T) {
	var calls int32
	current := time.Unix(0, 0)
	clock := func() time.Time { return current }

	c := New(time.Minute, func(key string) int {
		atomic.AddInt32(&calls, 1)
		return int(calls)
	}, clock)

	c.Get("a")               // load #1
	current = current.Add(2 * time.Minute) // past the TTL
	c.Get("a")               // load #2

	if calls != 2 {
		t.Errorf("loader called %d times, want 2 (value should expire)", calls)
	}
}

// TestGetLoadsOnceUnderConcurrency is the guarantee the naive lock-scope fix
// breaks: many goroutines racing on the same cold key must still trigger
// exactly one load. Run with -race to also verify the locking is sound.
func TestGetLoadsOnceUnderConcurrency(t *testing.T) {
	var calls int32
	c := New(time.Minute, func(key string) int {
		atomic.AddInt32(&calls, 1)
		time.Sleep(5 * time.Millisecond) // slow load
		return 7
	}, func() time.Time { return time.Unix(0, 0) })

	const n = 50
	var wg sync.WaitGroup
	results := make([]int, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = c.Get("a")
		}()
	}
	wg.Wait()

	if calls != 1 {
		t.Errorf("loader called %d times under concurrency, want exactly 1", calls)
	}
	for i, got := range results {
		if got != 7 {
			t.Errorf("goroutine %d got %d, want 7", i, got)
		}
	}
}

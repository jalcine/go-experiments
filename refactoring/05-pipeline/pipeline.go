// Package pipeline is refactoring exercise 05: concurrency (hard).
//
// GOAL
//
//	Fix the concurrency defects below while keeping the results correct.
//	Run the tests BOTH ways:
//
//	    go test ./...          # functional behavior (passes today)
//	    go test -race ./...    # data-race detector (FAILS today — that's the point)
//
//	Your refactor must make `go test -race ./...` pass and keep the ordinary
//	run green.
//
// SMELLS TO HUNT
//
//   - a data race: `processed` is incremented from every goroutine with no
//     synchronization. The results slice is fine (distinct indices), but the
//     counter is not — the race detector will point right at it.
//   - unbounded fan-out: one goroutine per input element. With a million
//     inputs that is a million goroutines. Bound the concurrency with a worker
//     pool (a semaphore channel, or N workers draining a jobs channel).
//   - no cancellation: there is no way to stop early. Thread a context.Context
//     through so a caller can cancel, and return ctx.Err() when it does.
//
// Suggested direction: keep ProcessAll's signature working for the existing
// tests, but consider adding a context-aware variant. For the counter, reach
// for sync/atomic or a mutex (or derive the count from len — do you even need
// the counter?).
package pipeline

import (
	"sync"
	"time"
)

// slowSquare simulates a unit of real work.
func slowSquare(n int) int {
	time.Sleep(time.Millisecond)
	return n * n
}

// ProcessAll squares every input concurrently and returns the results in the
// same order as the input.
func ProcessAll(nums []int) []int {
	results := make([]int, len(nums))
	processed := 0 // written from every goroutine without synchronization

	var wg sync.WaitGroup
	for i, n := range nums {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = slowSquare(n)
			processed++ // DATA RACE
		}()
	}
	wg.Wait()

	return results
}

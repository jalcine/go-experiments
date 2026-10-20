// Package stringutil is refactoring exercise 01: idiomatic cleanups (easy).
//
// GOAL
//
//	Make this code read like idiomatic Go. Do not change any observable
//	behavior: `go test ./...` from this directory must stay green the whole
//	time. Refactor in small steps and re-run the tests after each one.
//
// SMELLS TO HUNT
//
//   - `else` branches after a `return` (and needless nesting generally)
//   - redundant booleans: `if cond { return true } else { return false }`
//   - C-style index loops where `range` reads better
//   - building strings by repeated concatenation instead of a builder
//   - awkward empty/nil checks
//   - functions that reimplement what the standard library already offers
//
// Everything here is ASCII-only on purpose; you do not need to worry about
// multi-byte runes for this exercise.
package stringutil

import "strings"

// IsEmpty reports whether s has no characters.
func IsEmpty(s string) bool {
	return len(s) == 0
}

// Reverse returns s with its characters in reverse order.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CountVowels returns the number of ASCII vowels (either case) in s.
func CountVowels(s string) int {
	count := 0
	for _, c := range []rune(s) {
		if strings.ContainsRune("aeiouAEIOU", c) {
			count++
		}
	}
	return count
}

// IsPalindrome reports whether s reads the same forwards and backwards.
func IsPalindrome(s string) bool {
	return s == "" || s == Reverse(s)
}

// Truncate shortens s to at most max characters, appending "..." when it had
// to cut anything off.
func Truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

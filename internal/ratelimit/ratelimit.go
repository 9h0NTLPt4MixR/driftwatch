// Package ratelimit provides a simple token-bucket rate limiter for
// controlling how frequently drift scans are allowed to run against
// remote service endpoints.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// Limiter enforces a maximum number of operations per time window.
type Limiter struct {
	mu       sync.Mutex
	rate     int
	window   time.Duration
	buckets  map[string][]time.Time
}

// New creates a Limiter that allows at most rate calls per window per key.
func New(rate int, window time.Duration) (*Limiter, error) {
	if rate <= 0 {
		return nil, fmt.Errorf("ratelimit: rate must be positive, got %d", rate)
	}
	if window <= 0 {
		return nil, fmt.Errorf("ratelimit: window must be positive, got %s", window)
	}
	return &Limiter{
		rate:    rate,
		window:  window,
		buckets: make(map[string][]time.Time),
	}, nil
}

// Allow reports whether the operation identified by key is permitted.
// It records the attempt and evicts timestamps outside the current window.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	times := l.buckets[key]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.rate {
		l.buckets[key] = valid
		return false
	}

	l.buckets[key] = append(valid, now)
	return true
}

// Remaining returns how many calls are still permitted for key within
// the current window.
func (l *Limiter) Remaining(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	count := 0
	for _, t := range l.buckets[key] {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := l.rate - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

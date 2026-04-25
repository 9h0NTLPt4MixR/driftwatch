// Package throttle provides a configurable delay mechanism between
// service fetch operations to avoid overwhelming remote endpoints.
package throttle

import (
	"fmt"
	"time"
)

// Config holds throttle settings.
type Config struct {
	// Delay is the duration to wait between each fetch operation.
	Delay time.Duration
	// Burst allows this many operations before throttling begins.
	Burst int
}

// Throttle controls the rate of sequential operations.
type Throttle struct {
	delay   time.Duration
	burst   int
	count   int
	sleepFn func(time.Duration)
}

// New creates a new Throttle from the given Config.
// Returns an error if Delay is negative or Burst is less than 1.
func New(cfg Config) (*Throttle, error) {
	if cfg.Delay < 0 {
		return nil, fmt.Errorf("throttle: delay must be non-negative, got %s", cfg.Delay)
	}
	if cfg.Burst < 1 {
		return nil, fmt.Errorf("throttle: burst must be at least 1, got %d", cfg.Burst)
	}
	return &Throttle{
		delay:   cfg.Delay,
		burst:   cfg.Burst,
		sleepFn: time.Sleep,
	}, nil
}

// Wait should be called before each fetch operation. It sleeps for the
// configured delay once the burst allowance has been consumed.
func (t *Throttle) Wait() {
	t.count++
	if t.count > t.burst {
		t.sleepFn(t.delay)
	}
}

// Reset resets the internal operation counter, allowing another burst
// window to begin without sleeping.
func (t *Throttle) Reset() {
	t.count = 0
}

// Stats returns the current operation count and burst limit for
// observability or logging purposes.
func (t *Throttle) Stats() (count, burst int) {
	return t.count, t.burst
}

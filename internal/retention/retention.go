// Package retention provides policies for managing how long drift scan
// results are retained before being eligible for removal.
package retention

import (
	"errors"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Policy defines how results should be retained.
type Policy struct {
	// MaxAge is the maximum age of a result before it is considered expired.
	MaxAge time.Duration
	// MaxEntries is the maximum number of results to retain across all services.
	// Zero means unlimited.
	MaxEntries int
	// KeepDrifted ensures drifted results are never removed regardless of age.
	KeepDrifted bool
}

// Result holds the output of an Apply call.
type Result struct {
	Retained []drift.Result
	Removed  int
}

// Apply filters results according to the given Policy.
// It returns an error if the policy is invalid.
func Apply(results []drift.Result, p Policy) (Result, error) {
	if p.MaxAge < 0 {
		return Result{}, errors.New("retention: MaxAge must be non-negative")
	}
	if p.MaxEntries < 0 {
		return Result{}, errors.New("retention: MaxEntries must be non-negative")
	}

	now := time.Now()
	var retained []drift.Result

	for _, r := range results {
		if p.KeepDrifted && r.HasDrift {
			retained = append(retained, r)
			continue
		}
		if p.MaxAge > 0 && !r.Timestamp.IsZero() {
			if now.Sub(r.Timestamp) > p.MaxAge {
				continue
			}
		}
		retained = append(retained, r)
	}

	if p.MaxEntries > 0 && len(retained) > p.MaxEntries {
		retained = retained[len(retained)-p.MaxEntries:]
	}

	return Result{
		Retained: retained,
		Removed:  len(results) - len(retained),
	}, nil
}

// Stats returns a human-readable summary of a retention result.
func Stats(r Result) map[string]int {
	return map[string]int{
		"retained": len(r.Retained),
		"removed":  r.Removed,
	}
}

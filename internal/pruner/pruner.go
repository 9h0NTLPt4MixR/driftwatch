// Package pruner removes stale or irrelevant drift results based on age,
// status, or explicit exclusion criteria.
package pruner

import (
	"time"

	"github.com/driftwatch/internal/drift"
)

// Options controls which results are pruned.
type Options struct {
	// MaxAge removes results older than this duration. Zero means no age filter.
	MaxAge time.Duration

	// RemoveClean drops results with no drift.
	RemoveClean bool

	// ExcludeServices is a set of service names to drop entirely.
	ExcludeServices []string
}

// Prune filters results according to opts and returns the surviving slice.
// Results without a timestamp are treated as current when MaxAge is set.
func Prune(results []drift.Result, opts Options) []drift.Result {
	excluded := make(map[string]bool, len(opts.ExcludeServices))
	for _, s := range opts.ExcludeServices {
		excluded[s] = true
	}

	cutoff := time.Time{}
	if opts.MaxAge > 0 {
		cutoff = time.Now().Add(-opts.MaxAge)
	}

	out := make([]drift.Result, 0, len(results))
	for _, r := range results {
		if excluded[r.Service] {
			continue
		}
		if opts.RemoveClean && !r.HasDrift {
			continue
		}
		if !cutoff.IsZero() && !r.CheckedAt.IsZero() && r.CheckedAt.Before(cutoff) {
			continue
		}
		out = append(out, r)
	}
	return out
}

// Stats holds a summary of what Prune removed.
type Stats struct {
	Total    int
	Retained int
	Removed  int
}

// PruneWithStats is like Prune but also returns removal statistics.
func PruneWithStats(results []drift.Result, opts Options) ([]drift.Result, Stats) {
	kept := Prune(results, opts)
	return kept, Stats{
		Total:    len(results),
		Retained: len(kept),
		Removed:  len(results) - len(kept),
	}
}

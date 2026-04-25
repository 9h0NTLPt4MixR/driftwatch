// Package trimmer removes stale or redundant drift results based on
// configurable retention rules such as max age and max results per service.
package trimmer

import (
	"time"

	"github.com/driftwatch/internal/drift"
)

// Options controls how trimming is applied.
type Options struct {
	// MaxAge removes results older than this duration. Zero means no age limit.
	MaxAge time.Duration
	// MaxPerService keeps at most this many results per service (newest first).
	// Zero means no limit.
	MaxPerService int
	// OnlyDrifted drops all clean (non-drifted) results when true.
	OnlyDrifted bool
}

// Result holds trimming statistics alongside the trimmed results.
type Result struct {
	Kept    []drift.Result
	Removed int
}

// Trim applies the given Options to results and returns a Result containing
// the kept entries and a count of how many were removed.
func Trim(results []drift.Result, opts Options) Result {
	now := time.Now()
	counts := make(map[string]int)
	kept := make([]drift.Result, 0, len(results))
	removed := 0

	for _, r := range results {
		if opts.OnlyDrifted && !r.Drifted {
			removed++
			continue
		}

		if opts.MaxAge > 0 && !r.Timestamp.IsZero() {
			if now.Sub(r.Timestamp) > opts.MaxAge {
				removed++
				continue
			}
		}

		if opts.MaxPerService > 0 {
			if counts[r.Service] >= opts.MaxPerService {
				removed++
				continue
			}
		}

		counts[r.Service]++
		kept = append(kept, r)
	}

	return Result{Kept: kept, Removed: removed}
}

// Package truncator limits the number of diffs per result and the total number
// of results returned, useful for capping output in large deployments.
package truncator

import (
	"fmt"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Options controls how truncation is applied.
type Options struct {
	// MaxResults caps the total number of ScanResult entries returned.
	// Zero means no limit.
	MaxResults int

	// MaxDiffsPerResult caps the number of Diff entries kept per ScanResult.
	// Zero means no limit.
	MaxDiffsPerResult int

	// OnlyDrifted drops results with no diffs before applying MaxResults.
	OnlyDrifted bool
}

// Stats describes what was removed during truncation.
type Stats struct {
	ResultsDropped int
	DiffsDropped   int
}

// Apply truncates results according to opts and returns the trimmed slice
// together with a Stats summary.
func Apply(results []drift.ScanResult, opts Options) ([]drift.ScanResult, Stats) {
	var stats Stats
	out := make([]drift.ScanResult, 0, len(results))

	for _, r := range results {
		if opts.OnlyDrifted && len(r.Diffs) == 0 {
			stats.ResultsDropped++
			continue
		}

		if opts.MaxDiffsPerResult > 0 && len(r.Diffs) > opts.MaxDiffsPerResult {
			stats.DiffsDropped += len(r.Diffs) - opts.MaxDiffsPerResult
			r.Diffs = r.Diffs[:opts.MaxDiffsPerResult]
		}

		out = append(out, r)
	}

	if opts.MaxResults > 0 && len(out) > opts.MaxResults {
		stats.ResultsDropped += len(out) - opts.MaxResults
		out = out[:opts.MaxResults]
	}

	return out, stats
}

// Summary returns a human-readable description of the truncation stats.
func Summary(s Stats, at time.Time) string {
	return fmt.Sprintf("[%s] truncator: dropped %d result(s), %d diff(s)",
		at.UTC().Format(time.RFC3339), s.ResultsDropped, s.DiffsDropped)
}

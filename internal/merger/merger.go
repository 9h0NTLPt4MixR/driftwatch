// Package merger combines multiple drift scan results from different sources
// into a unified result set, resolving conflicts by preferring the most recent entry.
package merger

import (
	"sort"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Options controls how results are merged.
type Options struct {
	// PreferLatest keeps the most recently scanned result when duplicates exist.
	PreferLatest bool
	// DeduplicateKeys removes duplicate diff entries within a single result.
	DeduplicateKeys bool
}

// Merge combines multiple slices of ScanResult into one, resolving duplicates
// by service name according to the provided Options.
func Merge(batches [][]drift.ScanResult, opts Options) []drift.ScanResult {
	if len(batches) == 0 {
		return nil
	}

	index := make(map[string]drift.ScanResult)

	for _, batch := range batches {
		for _, r := range batch {
			existing, found := index[r.Service]
			if !found {
				if opts.DeduplicateKeys {
					r.Diffs = dedupeDiffs(r.Diffs)
				}
				index[r.Service] = r
				continue
			}
			if opts.PreferLatest && r.ScannedAt.After(existing.ScannedAt) {
				if opts.DeduplicateKeys {
					r.Diffs = dedupeDiffs(r.Diffs)
				}
				index[r.Service] = r
			} else if !opts.PreferLatest {
				merged := existing
				merged.Diffs = append(merged.Diffs, r.Diffs...)
				if opts.DeduplicateKeys {
					merged.Diffs = dedupeDiffs(merged.Diffs)
				}
				merged.Drifted = len(merged.Diffs) > 0
				if r.ScannedAt.After(merged.ScannedAt) {
					merged.ScannedAt = r.ScannedAt
				}
				index[r.Service] = merged
			}
		}
	}

	results := make([]drift.ScanResult, 0, len(index))
	for _, r := range index {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Service < results[j].Service
	})
	return results
}

// dedupeDiffs removes duplicate Diff entries by key.
func dedupeDiffs(diffs []drift.Diff) []drift.Diff {
	seen := make(map[string]struct{})
	out := diffs[:0:0]
	for _, d := range diffs {
		if _, ok := seen[d.Key]; !ok {
			seen[d.Key] = struct{}{}
			out = append(out, d)
		}
	}
	return out
}

// sentinel to avoid unused import
var _ = time.Now

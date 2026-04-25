// Package aggregator merges drift results from multiple scans into a
// unified view, de-duplicating services and combining per-key diffs.
package aggregator

import (
	"fmt"
	"sort"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Options controls how results are merged.
type Options struct {
	// PreferLatest, when true, overwrites an existing service entry with the
	// newer result instead of merging diffs.
	PreferLatest bool
}

// Merge combines multiple slices of drift.Result into a single deduplicated
// slice. When two results share the same ServiceName, their Diffs are unioned
// by key (last-write wins per key) unless PreferLatest is set, in which case
// the whole result is replaced.
func Merge(batches [][]drift.Result, opts Options) []drift.Result {
	index := make(map[string]*drift.Result)
	order := []string{}

	for _, batch := range batches {
		for i := range batch {
			r := batch[i]
			existing, found := index[r.ServiceName]
			if !found {
				copy := r
				index[r.ServiceName] = &copy
				order = append(order, r.ServiceName)
				continue
			}
			if opts.PreferLatest {
				copy := r
				index[r.ServiceName] = &copy
				continue
			}
			// Merge diffs by key; incoming value wins.
			diffMap := diffsByKey(existing.Diffs)
			for _, d := range r.Diffs {
				diffMap[d.Key] = d
			}
			existing.Diffs = flattenDiffs(diffMap)
			existing.Drifted = len(existing.Diffs) > 0
		}
	}

	results := make([]drift.Result, 0, len(order))
	for _, name := range order {
		results = append(results, *index[name])
	}
	return results
}

// Summary returns a human-readable one-liner describing the merged set.
func Summary(results []drift.Result) string {
	total := len(results)
	drifted := 0
	for _, r := range results {
		if r.Drifted {
			drifted++
		}
	}
	return fmt.Sprintf("%d service(s) aggregated, %d drifted", total, drifted)
}

func diffsByKey(diffs []drift.Diff) map[string]drift.Diff {
	m := make(map[string]drift.Diff, len(diffs))
	for _, d := range diffs {
		m[d.Key] = d
	}
	return m
}

func flattenDiffs(m map[string]drift.Diff) []drift.Diff {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]drift.Diff, 0, len(keys))
	for _, k := range keys {
		out = append(out, m[k])
	}
	return out
}

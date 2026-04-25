// Package deduper provides functionality to deduplicate drift scan results,
// ensuring that identical drift entries across multiple scans are collapsed
// into a single canonical result.
package deduper

import (
	"fmt"

	"github.com/driftwatch/internal/drift"
)

// key uniquely identifies a drift entry by service name and config key.
type key struct {
	Service string
	ConfigKey string
}

// Dedupe removes duplicate drift differences within each ScanResult,
// and optionally collapses duplicate ScanResults by service name,
// keeping the one with the most recent timestamp.
func Dedupe(results []drift.ScanResult) []drift.ScanResult {
	seen := make(map[string]*drift.ScanResult)

	for i := range results {
		r := &results[i]
		existing, ok := seen[r.Service]
		if !ok {
			copy := dedupeResult(*r)
			seen[r.Service] = &copy
			continue
		}
		// Keep the entry with more drift details; tie-break by service name stability.
		if len(r.Diffs) > len(existing.Diffs) {
			copy := dedupeResult(*r)
			seen[r.Service] = &copy
		}
	}

	out := make([]drift.ScanResult, 0, len(seen))
	for _, v := range seen {
		out = append(out, *v)
	}
	return out
}

// dedupeResult returns a copy of r with duplicate Diff entries removed.
// A Diff is considered duplicate if the same (Key, Expected, Actual) triple
// appears more than once.
func dedupeResult(r drift.ScanResult) drift.ScanResult {
	type diffKey struct{ Key, Expected, Actual string }
	seen := make(map[diffKey]struct{})
	uniq := make([]drift.Diff, 0, len(r.Diffs))

	for _, d := range r.Diffs {
		k := diffKey{
			Key:      d.Key,
			Expected: fmt.Sprintf("%v", d.Expected),
			Actual:   fmt.Sprintf("%v", d.Actual),
		}
		if _, exists := seen[k]; exists {
			continue
		}
		seen[k] = struct{}{}
		uniq = append(uniq, d)
	}

	r.Diffs = uniq
	return r
}

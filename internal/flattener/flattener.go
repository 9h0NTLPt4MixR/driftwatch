// Package flattener collapses nested drift results into a flat list of
// key-value diff entries, suitable for tabular output or further processing.
package flattener

import (
	"fmt"
	"sort"

	"github.com/driftwatch/internal/drift"
)

// Entry represents a single flattened diff row.
type Entry struct {
	Service  string
	Key      string
	Expected string
	Actual   string
	Drifted  bool
	ScannedAt string
}

// Flatten converts a slice of drift.Result into a flat list of Entry values.
// Each diff within a result becomes its own Entry. Results with no diffs are
// included as a single clean Entry when includeClean is true.
func Flatten(results []drift.Result, includeClean bool) []Entry {
	var entries []Entry

	for _, r := range results {
		if len(r.Diffs) == 0 {
			if includeClean {
				entries = append(entries, Entry{
					Service:   r.Service,
					Drifted:   false,
					ScannedAt: formatTime(r),
				})
			}
			continue
		}
		for _, d := range r.Diffs {
			entries = append(entries, Entry{
				Service:   r.Service,
				Key:       d.Key,
				Expected:  fmt.Sprintf("%v", d.Expected),
				Actual:    fmt.Sprintf("%v", d.Actual),
				Drifted:   true,
				ScannedAt: formatTime(r),
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Service != entries[j].Service {
			return entries[i].Service < entries[j].Service
		}
		return entries[i].Key < entries[j].Key
	})

	return entries
}

func formatTime(r drift.Result) string {
	if r.ScannedAt.IsZero() {
		return ""
	}
	return r.ScannedAt.Format("2006-01-02T15:04:05Z")
}

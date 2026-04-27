// Package sorter provides utilities for ordering drift results
// by various fields such as service name, drift count, or scan time.
package sorter

import (
	"sort"
	"strings"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Field represents a sortable attribute of a drift result.
type Field string

const (
	ByService   Field = "service"
	ByDriftCount Field = "drift_count"
	ByScannedAt  Field = "scanned_at"
)

// Options controls how results are sorted.
type Options struct {
	By        Field
	Descending bool
}

// Sort returns a new slice of results ordered according to opts.
// The original slice is not modified.
func Sort(results []drift.Result, opts Options) []drift.Result {
	out := make([]drift.Result, len(results))
	copy(out, results)

	sort.SliceStable(out, func(i, j int) bool {
		less := less(out[i], out[j], opts.By)
		if opts.Descending {
			return !less
		}
		return less
	})

	return out
}

func less(a, b drift.Result, by Field) bool {
	switch by {
	case ByDriftCount:
		return len(a.Diffs) < len(b.Diffs)
	case ByScannedAt:
		at := parseTime(a.ScannedAt)
		bt := parseTime(b.ScannedAt)
		return at.Before(bt)
	default: // ByService
		return strings.ToLower(a.Service) < strings.ToLower(b.Service)
	}
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

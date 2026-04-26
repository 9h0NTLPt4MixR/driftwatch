// Package rollup aggregates drift results into time-bucketed summaries
// suitable for trend analysis and reporting over arbitrary time windows.
package rollup

import (
	"sort"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Granularity defines the time bucket size for rollup.
type Granularity string

const (
	Hourly  Granularity = "hourly"
	Daily   Granularity = "daily"
	Weekly  Granularity = "weekly"
)

// Bucket holds aggregated drift statistics for a single time window.
type Bucket struct {
	WindowStart   time.Time `json:"window_start"`
	WindowEnd     time.Time `json:"window_end"`
	TotalServices int       `json:"total_services"`
	DriftedCount  int       `json:"drifted_count"`
	CleanCount    int       `json:"clean_count"`
	TotalDiffs    int       `json:"total_diffs"`
	DriftRate     float64   `json:"drift_rate"`
}

// Compute groups the provided results by the given granularity and returns
// a slice of Buckets sorted chronologically.
func Compute(results []drift.Result, gran Granularity) []Bucket {
	if len(results) == 0 {
		return nil
	}

	buckets := make(map[time.Time]*Bucket)

	for _, r := range results {
		key := bucketKey(r.Timestamp, gran)
		b, ok := buckets[key]
		if !ok {
			end := bucketEnd(key, gran)
			b = &Bucket{WindowStart: key, WindowEnd: end}
			buckets[key] = b
		}
		b.TotalServices++
		if r.HasDrift {
			b.DriftedCount++
			b.TotalDiffs += len(r.Diffs)
		} else {
			b.CleanCount++
		}
	}

	out := make([]Bucket, 0, len(buckets))
	for _, b := range buckets {
		if b.TotalServices > 0 {
			b.DriftRate = float64(b.DriftedCount) / float64(b.TotalServices)
		}
		out = append(out, *b)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].WindowStart.Before(out[j].WindowStart)
	})
	return out
}

func bucketKey(t time.Time, gran Granularity) time.Time {
	switch gran {
	case Hourly:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case Weekly:
		weekday := int(t.Weekday())
		start := t.AddDate(0, 0, -weekday)
		return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, t.Location())
	default: // Daily
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}
}

func bucketEnd(start time.Time, gran Granularity) time.Time {
	switch gran {
	case Hourly:
		return start.Add(time.Hour)
	case Weekly:
		return start.AddDate(0, 0, 7)
	default:
		return start.AddDate(0, 0, 1)
	}
}

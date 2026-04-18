// Package summary provides aggregation of drift scan results into
// a high-level statistical overview.
package summary

import "github.com/driftwatch/internal/drift"

// Stats holds aggregated metrics from a set of scan results.
type Stats struct {
	TotalServices  int            `json:"total_services"`
	DriftedCount   int            `json:"drifted_count"`
	CleanCount     int            `json:"clean_count"`
	TotalMismatches int           `json:"total_mismatches"`
	ByService      map[string]int `json:"by_service"`
}

// Compute aggregates drift results into a Stats summary.
func Compute(results []drift.Result) Stats {
	s := Stats{
		ByService: make(map[string]int),
	}

	for _, r := range results {
		s.TotalServices++
		mismatches := len(r.Diffs)
		s.TotalMismatches += mismatches
		if mismatches > 0 {
			s.DriftedCount++
		} else {
			s.CleanCount++
		}
		s.ByService[r.Service] = mismatches
	}

	return s
}

// DriftRate returns the percentage of services with drift (0–100).
func (s Stats) DriftRate() float64 {
	if s.TotalServices == 0 {
		return 0
	}
	return float64(s.DriftedCount) / float64(s.TotalServices) * 100
}

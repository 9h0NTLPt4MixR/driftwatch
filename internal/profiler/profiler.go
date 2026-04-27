// Package profiler measures per-service scan latency and aggregates timing
// statistics across a drift scan run.
package profiler

import (
	"fmt"
	"sort"
	"time"
)

// Entry holds timing data for a single service scan.
type Entry struct {
	Service  string        `json:"service"`
	Duration time.Duration `json:"duration_ns"`
	Drifted  bool          `json:"drifted"`
}

// Stats summarises timing across all recorded entries.
type Stats struct {
	Total   int           `json:"total"`
	Min     time.Duration `json:"min_ns"`
	Max     time.Duration `json:"max_ns"`
	Avg     time.Duration `json:"avg_ns"`
	Entries []Entry       `json:"entries"`
}

// Profiler records scan durations for individual services.
type Profiler struct {
	entries []Entry
}

// New returns an initialised Profiler.
func New() *Profiler {
	return &Profiler{}
}

// Record stores the elapsed duration for a named service.
func (p *Profiler) Record(service string, d time.Duration, drifted bool) {
	p.entries = append(p.entries, Entry{
		Service:  service,
		Duration: d,
		Drifted:  drifted,
	})
}

// Compute returns aggregated Stats over all recorded entries.
// Returns an error if no entries have been recorded.
func (p *Profiler) Compute() (Stats, error) {
	if len(p.entries) == 0 {
		return Stats{}, fmt.Errorf("profiler: no entries recorded")
	}

	min := p.entries[0].Duration
	max := p.entries[0].Duration
	var total time.Duration

	for _, e := range p.entries {
		if e.Duration < min {
			min = e.Duration
		}
		if e.Duration > max {
			max = e.Duration
		}
		total += e.Duration
	}

	// Return entries sorted slowest-first for easy inspection.
	sorted := make([]Entry, len(p.entries))
	copy(sorted, p.entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Duration > sorted[j].Duration
	})

	return Stats{
		Total:   len(p.entries),
		Min:     min,
		Max:     max,
		Avg:     total / time.Duration(len(p.entries)),
		Entries: sorted,
	}, nil
}

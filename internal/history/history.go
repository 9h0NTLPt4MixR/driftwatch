// Package history tracks scan results over time, enabling trend analysis.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Entry represents a single historical scan record.
type Entry struct {
	Timestamp time.Time     `json:"timestamp"`
	Results   []drift.Result `json:"results"`
}

// Record appends a new scan entry to the history file.
func Record(path string, results []drift.Result) error {
	entries, err := LoadAll(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("history: load: %w", err)
	}
	entries = append(entries, Entry{
		Timestamp: time.Now().UTC(),
		Results:   results,
	})
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("history: write: %w", err)
	}
	return nil
}

// LoadAll reads all history entries from the given file.
func LoadAll(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("history: unmarshal: %w", err)
	}
	return entries, nil
}

// DriftTrend returns the count of drifted services per entry, sorted by time.
func DriftTrend(entries []Entry) []TrendPoint {
	sorted := make([]Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})
	points := make([]TrendPoint, 0, len(sorted))
	for _, e := range sorted {
		count := 0
		for _, r := range e.Results {
			if r.HasDrift {
				count++
			}
		}
		points = append(points, TrendPoint{Timestamp: e.Timestamp, DriftedCount: count})
	}
	return points
}

// TrendPoint holds drift count at a point in time.
type TrendPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	DriftedCount int       `json:"drifted_count"`
}

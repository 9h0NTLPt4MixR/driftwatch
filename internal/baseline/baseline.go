package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Entry represents a saved baseline for a single service.
type Entry struct {
	ServiceName string            `json:"service_name"`
	Expected    map[string]string `json:"expected"`
	Actual      map[string]string `json:"actual"`
	RecordedAt  time.Time         `json:"recorded_at"`
}

// File is the top-level structure persisted to disk.
type File struct {
	Version   int     `json:"version"`
	Baselines []Entry `json:"baselines"`
}

// Save writes the current scan results as a baseline to the given path.
func Save(path string, results []drift.Result) error {
	entries := make([]Entry, 0, len(results))
	for _, r := range results {
		entries = append(entries, Entry{
			ServiceName: r.ServiceName,
			Expected:    r.Expected,
			Actual:      r.Actual,
			RecordedAt:  time.Now().UTC(),
		})
	}

	f := File{Version: 1, Baselines: entries}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("baseline: write %s: %w", path, err)
	}
	return nil
}

// Load reads a baseline file from disk.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("baseline: file not found: %s", path)
		}
		return nil, fmt.Errorf("baseline: read %s: %w", path, err)
	}

	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("baseline: unmarshal: %w", err)
	}
	return &f, nil
}

// Compare returns new drift results relative to the saved baseline.
// Only keys present in the baseline expected map are compared.
// Services present in current but not in the baseline are included as-is.
func Compare(base *File, current []drift.Result) []drift.Result {
	index := make(map[string]Entry, len(base.Baselines))
	for _, e := range base.Baselines {
		index[e.ServiceName] = e
	}

	var out []drift.Result
	for _, r := range current {
		entry, ok := index[r.ServiceName]
		if !ok {
			out = append(out, r)
			continue
		}
		diffs := map[string]drift.Diff{}
		for k, want := range entry.Expected {
			got, exists := r.Actual[k]
			if !exists || got != want {
				diffs[k] = drift.Diff{Expected: want, Actual: got}
			}
		}
		out = append(out, drift.Result{
			ServiceName: r.ServiceName,
			Expected:    entry.Expected,
			Actual:      r.Actual,
			Diffs:       diffs,
		})
	}
	return out
}

// HasDrift reports whether any result in the given slice contains at least one diff.
func HasDrift(results []drift.Result) bool {
	for _, r := range results {
		if len(r.Diffs) > 0 {
			return true
		}
	}
	return false
}

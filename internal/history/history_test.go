package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/history"
)

func makeResults(hasDrift ...bool) []drift.Result {
	results := make([]drift.Result, len(hasDrift))
	for i, d := range hasDrift {
		results[i] = drift.Result{Service: "svc", HasDrift: d}
	}
	return results
}

func TestRecord_And_LoadAll_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	results := makeResults(true, false)
	if err := history.Record(path, results); err != nil {
		t.Fatalf("Record: %v", err)
	}
	entries, err := history.LoadAll(path)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if len(entries[0].Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(entries[0].Results))
	}
}

func TestRecord_Appends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	for i := 0; i < 3; i++ {
		if err := history.Record(path, makeResults(i%2 == 0)); err != nil {
			t.Fatalf("Record iteration %d: %v", i, err)
		}
	}
	entries, _ := history.LoadAll(path)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestLoadAll_MissingFile(t *testing.T) {
	_, err := history.LoadAll("/nonexistent/history.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadAll_InvalidJSON(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.json")
	f.WriteString("not json")
	f.Close()
	_, err := history.LoadAll(f.Name())
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDriftTrend(t *testing.T) {
	now := time.Now().UTC()
	entries := []history.Entry{
		{Timestamp: now.Add(-2 * time.Hour), Results: makeResults(true, true)},
		{Timestamp: now.Add(-1 * time.Hour), Results: makeResults(false, false)},
		{Timestamp: now, Results: makeResults(true)},
	}
	points := history.DriftTrend(entries)
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}
	if points[0].DriftedCount != 2 {
		t.Errorf("expected 2, got %d", points[0].DriftedCount)
	}
	if points[1].DriftedCount != 0 {
		t.Errorf("expected 0, got %d", points[1].DriftedCount)
	}
	if points[2].DriftedCount != 1 {
		t.Errorf("expected 1, got %d", points[2].DriftedCount)
	}
}

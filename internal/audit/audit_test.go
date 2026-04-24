package audit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/audit"
	"github.com/driftwatch/internal/drift"
)

func makeResults(drifted bool) []drift.Result {
	return []drift.Result{
		{
			Service:  "auth-service",
			HasDrift: drifted,
			Diffs:    map[string]drift.Diff{},
		},
		{
			Service:  "payment-service",
			HasDrift: false,
			Diffs:    map[string]drift.Diff{},
		},
	}
}

func tmpPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.ndjson")
}

func TestRecord_And_LoadAll_RoundTrip(t *testing.T) {
	path := tmpPath(t)
	results := makeResults(true)

	if err := audit.Record(path, "ci-bot", "scan", results); err != nil {
		t.Fatalf("Record: %v", err)
	}

	entries, err := audit.LoadAll(path)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.User != "ci-bot" {
		t.Errorf("user: got %q, want %q", e.User, "ci-bot")
	}
	if e.Command != "scan" {
		t.Errorf("command: got %q, want %q", e.Command, "scan")
	}
	if e.TotalSvcs != 2 {
		t.Errorf("total_services: got %d, want 2", e.TotalSvcs)
	}
	if e.DriftCount != 1 {
		t.Errorf("drift_count: got %d, want 1", e.DriftCount)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestRecord_Appends(t *testing.T) {
	path := tmpPath(t)
	results := makeResults(false)

	for i := 0; i < 3; i++ {
		if err := audit.Record(path, "admin", "baseline", results); err != nil {
			t.Fatalf("Record iteration %d: %v", i, err)
		}
	}

	entries, err := audit.LoadAll(path)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestLoadAll_MissingFile(t *testing.T) {
	entries, err := audit.LoadAll("/nonexistent/audit.ndjson")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

func TestRecord_TimestampIsUTC(t *testing.T) {
	path := tmpPath(t)
	before := time.Now().UTC()

	if err := audit.Record(path, "dev", "scan", makeResults(false)); err != nil {
		t.Fatalf("Record: %v", err)
	}

	after := time.Now().UTC()
	entries, _ := audit.LoadAll(path)
	ts := entries[0].Timestamp

	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
	if ts.Location() != time.UTC {
		t.Errorf("expected UTC location, got %v", ts.Location())
	}
}

func TestRecord_CreatesFileIfAbsent(t *testing.T) {
	path := tmpPath(t)
	if err := audit.Record(path, "bot", "policy", makeResults(false)); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}

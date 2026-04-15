package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/snapshot"
)

func makeResults() []drift.Result {
	return []drift.Result{
		{
			Service: "api",
			Diffs: []drift.Diff{
				{Key: "LOG_LEVEL", Expected: "info", Actual: "debug"},
			},
		},
	}
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	results := makeResults()

	path, err := snapshot.Save(results, dir)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", path)
	}

	if filepath.Dir(path) != dir {
		t.Errorf("expected file in %s, got %s", dir, filepath.Dir(path))
	}
}

func TestSave_Load_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	results := makeResults()

	path, err := snapshot.Save(results, dir)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	entry, err := snapshot.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(entry.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(entry.Results))
	}
	if entry.Results[0].Service != "api" {
		t.Errorf("expected service 'api', got %q", entry.Results[0].Service)
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if entry.Timestamp.Location() != time.UTC {
		t.Error("expected UTC timestamp")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path/snap.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "snap*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("not-json")
	f.Close()

	_, err = snapshot.Load(f.Name())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

package baseline_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/driftwatch/internal/baseline"
	"github.com/driftwatch/internal/drift"
)

func makeResults() []drift.Result {
	return []drift.Result{
		{
			ServiceName: "api",
			Expected:    map[string]string{"LOG_LEVEL": "info", "PORT": "8080"},
			Actual:      map[string]string{"LOG_LEVEL": "debug", "PORT": "8080"},
			Diffs:       map[string]drift.Diff{"LOG_LEVEL": {Expected: "info", Actual: "debug"}},
		},
	}
}

func TestSave_And_Load_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	results := makeResults()
	if err := baseline.Save(path, results); err != nil {
		t.Fatalf("Save: %v", err)
	}

	f, err := baseline.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if f.Version != 1 {
		t.Errorf("version = %d, want 1", f.Version)
	}
	if len(f.Baselines) != 1 {
		t.Fatalf("baselines len = %d, want 1", len(f.Baselines))
	}
	e := f.Baselines[0]
	if e.ServiceName != "api" {
		t.Errorf("service = %q, want \"api\"", e.ServiceName)
	}
	if e.RecordedAt.IsZero() {
		t.Error("recorded_at should not be zero")
	}
	if e.RecordedAt.After(time.Now().Add(time.Second)) {
		t.Error("recorded_at is in the future")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := baseline.Load("/nonexistent/baseline.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("{invalid"), 0o644)

	_, err := baseline.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCompare_DetectsDrift(t *testing.T) {
	base := &baseline.File{
		Version: 1,
		Baselines: []baseline.Entry{
			{
				ServiceName: "api",
				Expected:    map[string]string{"LOG_LEVEL": "info", "PORT": "8080"},
				Actual:      map[string]string{"LOG_LEVEL": "info", "PORT": "8080"},
			},
		},
	}

	current := []drift.Result{
		{
			ServiceName: "api",
			Actual:      map[string]string{"LOG_LEVEL": "warn", "PORT": "8080"},
		},
	}

	out := baseline.Compare(base, current)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if _, hasDiff := out[0].Diffs["LOG_LEVEL"]; !hasDiff {
		t.Error("expected drift on LOG_LEVEL")
	}
	if _, hasDiff := out[0].Diffs["PORT"]; hasDiff {
		t.Error("PORT should not have drift")
	}
}

func TestCompare_NoDrift(t *testing.T) {
	base := &baseline.File{
		Version: 1,
		Baselines: []baseline.Entry{
			{ServiceName: "svc", Expected: map[string]string{"KEY": "val"}},
		},
	}
	current := []drift.Result{
		{ServiceName: "svc", Actual: map[string]string{"KEY": "val"}},
	}
	out := baseline.Compare(base, current)
	if len(out[0].Diffs) != 0 {
		t.Errorf("expected no diffs, got %v", out[0].Diffs)
	}
}

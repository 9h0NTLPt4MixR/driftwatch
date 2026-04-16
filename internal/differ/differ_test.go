package differ_test

import (
	"strings"
	"testing"

	"github.com/user/driftwatch/internal/differ"
	"github.com/user/driftwatch/internal/drift"
)

func makeResult(service string, drifts map[string]drift.Delta) drift.Result {
	return drift.Result{
		Service:  service,
		HasDrift: len(drifts) > 0,
		Diffs:    drifts,
	}
}

func TestCompute_NoDrift(t *testing.T) {
	results := []drift.Result{makeResult("svc", nil)}
	diffs := differ.Compute(results)
	if len(diffs) != 0 {
		t.Fatalf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestCompute_WithDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("svc", map[string]drift.Delta{
			"LOG_LEVEL": {Expected: "info", Actual: "debug"},
		}),
	}
	diffs := differ.Compute(results)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Lines[0].ChangeType != "changed" {
		t.Errorf("expected changed, got %s", diffs[0].Lines[0].ChangeType)
	}
}

func TestCompute_AddedKey(t *testing.T) {
	results := []drift.Result{
		makeResult("svc", map[string]drift.Delta{
			"NEW_KEY": {Expected: "", Actual: "value"},
		}),
	}
	diffs := differ.Compute(results)
	if diffs[0].Lines[0].ChangeType != "added" {
		t.Errorf("expected added, got %s", diffs[0].Lines[0].ChangeType)
	}
}

func TestCompute_RemovedKey(t *testing.T) {
	results := []drift.Result{
		makeResult("svc", map[string]drift.Delta{
			"OLD_KEY": {Expected: "val", Actual: ""},
		}),
	}
	diffs := differ.Compute(results)
	if diffs[0].Lines[0].ChangeType != "removed" {
		t.Errorf("expected removed, got %s", diffs[0].Lines[0].ChangeType)
	}
}

func TestFormat_Output(t *testing.T) {
	diffs := []differ.Diff{
		{
			Service: "api",
			Lines: []differ.DiffLine{
				{Key: "PORT", Old: "8080", New: "9090", ChangeType: "changed"},
				{Key: "DEBUG", Old: "", New: "true", ChangeType: "added"},
			},
		},
	}
	out := differ.Format(diffs)
	if !strings.Contains(out, "--- api") {
		t.Error("expected service header")
	}
	if !strings.Contains(out, "~ PORT") {
		t.Error("expected changed marker")
	}
	if !strings.Contains(out, "+ DEBUG") {
		t.Error("expected added marker")
	}
}

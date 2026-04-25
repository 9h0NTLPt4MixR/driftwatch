package deduper_test

import (
	"testing"

	"github.com/driftwatch/internal/deduper"
	"github.com/driftwatch/internal/drift"
)

func makeResult(service string, diffs []drift.Diff) drift.ScanResult {
	return drift.ScanResult{
		Service:  service,
		Drifted:  len(diffs) > 0,
		Diffs:    diffs,
	}
}

func TestDedupe_NoDuplicateServices(t *testing.T) {
	input := []drift.ScanResult{
		makeResult("alpha", nil),
		makeResult("beta", nil),
	}
	out := deduper.Dedupe(input)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestDedupe_CollapsesDuplicateService(t *testing.T) {
	diff := drift.Diff{Key: "PORT", Expected: "8080", Actual: "9090"}
	input := []drift.ScanResult{
		makeResult("alpha", nil),
		makeResult("alpha", []drift.Diff{diff}),
	}
	out := deduper.Dedupe(input)
	if len(out) != 1 {
		t.Fatalf("expected 1 result after dedup, got %d", len(out))
	}
	if len(out[0].Diffs) != 1 {
		t.Errorf("expected 1 diff in merged result, got %d", len(out[0].Diffs))
	}
}

func TestDedupe_RemovesDuplicateDiffs(t *testing.T) {
	diff := drift.Diff{Key: "LOG_LEVEL", Expected: "info", Actual: "debug"}
	input := []drift.ScanResult{
		makeResult("gamma", []drift.Diff{diff, diff, diff}),
	}
	out := deduper.Dedupe(input)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if len(out[0].Diffs) != 1 {
		t.Errorf("expected duplicate diffs collapsed to 1, got %d", len(out[0].Diffs))
	}
}

func TestDedupe_PreservesDistinctDiffs(t *testing.T) {
	input := []drift.ScanResult{
		makeResult("delta", []drift.Diff{
			{Key: "A", Expected: "1", Actual: "2"},
			{Key: "B", Expected: "x", Actual: "y"},
		}),
	}
	out := deduper.Dedupe(input)
	if len(out[0].Diffs) != 2 {
		t.Errorf("expected 2 distinct diffs preserved, got %d", len(out[0].Diffs))
	}
}

func TestDedupe_EmptyInput(t *testing.T) {
	out := deduper.Dedupe([]drift.ScanResult{})
	if len(out) != 0 {
		t.Errorf("expected empty output for empty input, got %d", len(out))
	}
}

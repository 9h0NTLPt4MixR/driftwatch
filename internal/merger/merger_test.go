package merger_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/merger"
)

func makeResult(service string, drifted bool, scannedAt time.Time, diffs ...drift.Diff) drift.ScanResult {
	return drift.ScanResult{
		Service:   service,
		Drifted:   drifted,
		ScannedAt: scannedAt,
		Diffs:     diffs,
	}
}

func TestMerge_EmptyBatches(t *testing.T) {
	result := merger.Merge(nil, merger.Options{})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMerge_SingleBatch_NoOverlap(t *testing.T) {
	now := time.Now()
	batch := []drift.ScanResult{
		makeResult("svc-a", false, now),
		makeResult("svc-b", true, now, drift.Diff{Key: "port", Expected: "8080", Actual: "9090"}),
	}
	out := merger.Merge([][]drift.ScanResult{batch}, merger.Options{})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestMerge_PreferLatest_Replaces(t *testing.T) {
	old := time.Now().Add(-1 * time.Hour)
	new := time.Now()

	batch1 := []drift.ScanResult{makeResult("svc-a", false, old)}
	batch2 := []drift.ScanResult{makeResult("svc-a", true, new, drift.Diff{Key: "env", Expected: "prod", Actual: "staging"})}

	out := merger.Merge([][]drift.ScanResult{batch1, batch2}, merger.Options{PreferLatest: true})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if !out[0].Drifted {
		t.Error("expected drifted=true from newer batch")
	}
}

func TestMerge_NoPreferLatest_MergesDiffs(t *testing.T) {
	now := time.Now()
	d1 := drift.Diff{Key: "port", Expected: "8080", Actual: "9090"}
	d2 := drift.Diff{Key: "env", Expected: "prod", Actual: "dev"}

	batch1 := []drift.ScanResult{makeResult("svc-a", true, now, d1)}
	batch2 := []drift.ScanResult{makeResult("svc-a", true, now, d2)}

	out := merger.Merge([][]drift.ScanResult{batch1, batch2}, merger.Options{PreferLatest: false})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if len(out[0].Diffs) != 2 {
		t.Errorf("expected 2 merged diffs, got %d", len(out[0].Diffs))
	}
}

func TestMerge_DeduplicateKeys_RemovesDuplicateDiffs(t *testing.T) {
	now := time.Now()
	d := drift.Diff{Key: "port", Expected: "8080", Actual: "9090"}

	batch1 := []drift.ScanResult{makeResult("svc-a", true, now, d)}
	batch2 := []drift.ScanResult{makeResult("svc-a", true, now, d)}

	out := merger.Merge([][]drift.ScanResult{batch1, batch2}, merger.Options{DeduplicateKeys: true})
	if len(out[0].Diffs) != 1 {
		t.Errorf("expected 1 deduped diff, got %d", len(out[0].Diffs))
	}
}

func TestMerge_SortsByServiceName(t *testing.T) {
	now := time.Now()
	batch := []drift.ScanResult{
		makeResult("zebra", false, now),
		makeResult("alpha", false, now),
		makeResult("mango", false, now),
	}
	out := merger.Merge([][]drift.ScanResult{batch}, merger.Options{})
	if out[0].Service != "alpha" || out[1].Service != "mango" || out[2].Service != "zebra" {
		t.Errorf("results not sorted: %v", out)
	}
}

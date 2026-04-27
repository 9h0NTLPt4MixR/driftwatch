package truncator_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/truncator"
)

func makeResult(service string, diffs int) drift.ScanResult {
	d := make([]drift.Diff, diffs)
	for i := range d {
		d[i] = drift.Diff{Key: fmt.Sprintf("key%d", i), Expected: "a", Actual: "b"}
	}
	return drift.ScanResult{
		Service:   service,
		Diffs:     d,
		ScannedAt: time.Now(),
	}
}

func TestApply_NoOptions_ReturnsAll(t *testing.T) {
	results := []drift.ScanResult{
		makeResult("svc-a", 3),
		makeResult("svc-b", 0),
	}
	out, stats := truncator.Apply(results, truncator.Options{})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if stats.ResultsDropped != 0 || stats.DiffsDropped != 0 {
		t.Fatalf("expected zero drops, got %+v", stats)
	}
}

func TestApply_OnlyDrifted_DropsClean(t *testing.T) {
	results := []drift.ScanResult{
		makeResult("svc-a", 2),
		makeResult("svc-b", 0),
		makeResult("svc-c", 1),
	}
	out, stats := truncator.Apply(results, truncator.Options{OnlyDrifted: true})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if stats.ResultsDropped != 1 {
		t.Fatalf("expected 1 dropped result, got %d", stats.ResultsDropped)
	}
}

func TestApply_MaxResults_Caps(t *testing.T) {
	results := []drift.ScanResult{
		makeResult("svc-a", 1),
		makeResult("svc-b", 1),
		makeResult("svc-c", 1),
	}
	out, stats := truncator.Apply(results, truncator.Options{MaxResults: 2})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if stats.ResultsDropped != 1 {
		t.Fatalf("expected 1 dropped result, got %d", stats.ResultsDropped)
	}
}

func TestApply_MaxDiffsPerResult_Caps(t *testing.T) {
	results := []drift.ScanResult{
		makeResult("svc-a", 5),
	}
	out, stats := truncator.Apply(results, truncator.Options{MaxDiffsPerResult: 2})
	if len(out[0].Diffs) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(out[0].Diffs))
	}
	if stats.DiffsDropped != 3 {
		t.Fatalf("expected 3 diffs dropped, got %d", stats.DiffsDropped)
	}
}

func TestApply_Combined_Options(t *testing.T) {
	results := []drift.ScanResult{
		makeResult("svc-a", 4),
		makeResult("svc-b", 0),
		makeResult("svc-c", 6),
		makeResult("svc-d", 2),
	}
	out, stats := truncator.Apply(results, truncator.Options{
		OnlyDrifted:       true,
		MaxResults:        2,
		MaxDiffsPerResult: 3,
	})
	// svc-b dropped (no diffs), then capped to 2 of remaining 3
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if stats.ResultsDropped != 2 { // svc-b + svc-d
		t.Fatalf("expected 2 results dropped, got %d", stats.ResultsDropped)
	}
	if stats.DiffsDropped != 3 { // svc-c capped from 6 to 3
		t.Fatalf("expected 3 diffs dropped, got %d", stats.DiffsDropped)
	}
}

func TestSummary_Format(t *testing.T) {
	at := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	s := truncator.Summary(truncator.Stats{ResultsDropped: 3, DiffsDropped: 7}, at)
	if s == "" {
		t.Fatal("expected non-empty summary")
	}
}

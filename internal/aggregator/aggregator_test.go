package aggregator_test

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/aggregator"
	"github.com/yourorg/driftwatch/internal/drift"
)

func makeResult(name string, drifted bool, keys ...string) drift.Result {
	diffs := make([]drift.Diff, 0, len(keys))
	for _, k := range keys {
		diffs = append(diffs, drift.Diff{Key: k, Expected: "a", Actual: "b"})
	}
	return drift.Result{ServiceName: name, Drifted: drifted, Diffs: diffs}
}

func TestMerge_EmptyBatches(t *testing.T) {
	out := aggregator.Merge(nil, aggregator.Options{})
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d", len(out))
	}
}

func TestMerge_SingleBatch_NoOverlap(t *testing.T) {
	batch := []drift.Result{
		makeResult("svc-a", true, "port"),
		makeResult("svc-b", false),
	}
	out := aggregator.Merge([][]drift.Result{batch}, aggregator.Options{})
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
}

func TestMerge_MergesDiffs(t *testing.T) {
	b1 := []drift.Result{makeResult("svc-a", true, "port")}
	b2 := []drift.Result{makeResult("svc-a", true, "timeout")}

	out := aggregator.Merge([][]drift.Result{b1, b2}, aggregator.Options{})
	if len(out) != 1 {
		t.Fatalf("expected 1 service, got %d", len(out))
	}
	if len(out[0].Diffs) != 2 {
		t.Fatalf("expected 2 merged diffs, got %d", len(out[0].Diffs))
	}
}

func TestMerge_PreferLatest_Replaces(t *testing.T) {
	b1 := []drift.Result{makeResult("svc-a", true, "port", "timeout")}
	b2 := []drift.Result{makeResult("svc-a", true, "replicas")}

	out := aggregator.Merge([][]drift.Result{b1, b2}, aggregator.Options{PreferLatest: true})
	if len(out) != 1 {
		t.Fatalf("expected 1 service, got %d", len(out))
	}
	// Only the latest result's single diff should survive.
	if len(out[0].Diffs) != 1 || out[0].Diffs[0].Key != "replicas" {
		t.Fatalf("expected only 'replicas' diff, got %+v", out[0].Diffs)
	}
}

func TestMerge_DedupesServices(t *testing.T) {
	b1 := []drift.Result{makeResult("svc-a", false), makeResult("svc-b", false)}
	b2 := []drift.Result{makeResult("svc-a", false)}

	out := aggregator.Merge([][]drift.Result{b1, b2}, aggregator.Options{})
	if len(out) != 2 {
		t.Fatalf("expected 2 unique services, got %d", len(out))
	}
}

func TestSummary(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, "port"),
		makeResult("svc-b", false),
		makeResult("svc-c", true, "timeout"),
	}
	got := aggregator.Summary(results)
	want := "3 service(s) aggregated, 2 drifted"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

package flattener_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/flattener"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service:   service,
		Drifted:   len(diffs) > 0,
		Diffs:     diffs,
		ScannedAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestFlatten_Empty(t *testing.T) {
	entries := flattener.Flatten(nil, true)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestFlatten_CleanResult_Included(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", nil)}
	entries := flattener.Flatten(results, true)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Drifted {
		t.Error("expected Drifted=false for clean result")
	}
	if entries[0].Service != "svc-a" {
		t.Errorf("unexpected service: %s", entries[0].Service)
	}
}

func TestFlatten_CleanResult_Excluded(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", nil)}
	entries := flattener.Flatten(results, false)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestFlatten_DriftedResult_ExpandsDiffs(t *testing.T) {
	diffs := []drift.Diff{
		{Key: "timeout", Expected: "30s", Actual: "60s"},
		{Key: "replicas", Expected: 2, Actual: 3},
	}
	results := []drift.Result{makeResult("svc-b", diffs)}
	entries := flattener.Flatten(results, false)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for _, e := range entries {
		if !e.Drifted {
			t.Error("expected Drifted=true")
		}
		if e.Service != "svc-b" {
			t.Errorf("unexpected service: %s", e.Service)
		}
	}
}

func TestFlatten_SortedByServiceThenKey(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-z", []drift.Diff{{Key: "b", Expected: "1", Actual: "2"}}),
		makeResult("svc-a", []drift.Diff{
			{Key: "z", Expected: "1", Actual: "2"},
			{Key: "a", Expected: "1", Actual: "2"},
		}),
	}
	entries := flattener.Flatten(results, false)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Service != "svc-a" || entries[0].Key != "a" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Key != "z" {
		t.Errorf("unexpected second entry key: %s", entries[1].Key)
	}
	if entries[2].Service != "svc-z" {
		t.Errorf("unexpected third entry service: %s", entries[2].Service)
	}
}

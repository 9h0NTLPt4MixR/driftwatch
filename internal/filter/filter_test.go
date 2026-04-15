package filter_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/filter"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service:  service,
		Diffs:    diffs,
		HasDrift: len(diffs) > 0,
	}
}

func TestApply_NoOptions(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", nil),
		makeResult("svc-b", []drift.Diff{{Key: "PORT", Expected: "8080", Actual: "9090"}}),
	}
	out := filter.Apply(results, filter.Options{})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestApply_OnlyDrifted(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", nil),
		makeResult("svc-b", []drift.Diff{{Key: "PORT", Expected: "8080", Actual: "9090"}}),
	}
	out := filter.Apply(results, filter.Options{OnlyDrifted: true})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Service != "svc-b" {
		t.Errorf("expected svc-b, got %s", out[0].Service)
	}
}

func TestApply_ServiceFilter(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", nil),
		makeResult("svc-b", nil),
		makeResult("svc-c", nil),
	}
	out := filter.Apply(results, filter.Options{Services: []string{"svc-a", "svc-c"}})
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestApply_KeyFilter(t *testing.T) {
	diffs := []drift.Diff{
		{Key: "PORT", Expected: "8080", Actual: "9090"},
		{Key: "LOG_LEVEL", Expected: "info", Actual: "debug"},
	}
	results := []drift.Result{makeResult("svc-a", diffs)}
	out := filter.Apply(results, filter.Options{Keys: []string{"PORT"}})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if len(out[0].Diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(out[0].Diffs))
	}
	if out[0].Diffs[0].Key != "PORT" {
		t.Errorf("expected PORT diff, got %s", out[0].Diffs[0].Key)
	}
}

func TestApply_KeyFilter_NoDriftAfterFilter(t *testing.T) {
	diffs := []drift.Diff{
		{Key: "LOG_LEVEL", Expected: "info", Actual: "debug"},
	}
	results := []drift.Result{makeResult("svc-a", diffs)}
	out := filter.Apply(results, filter.Options{OnlyDrifted: true, Keys: []string{"PORT"}})
	if len(out) != 0 {
		t.Fatalf("expected 0 results after key+drift filter, got %d", len(out))
	}
}

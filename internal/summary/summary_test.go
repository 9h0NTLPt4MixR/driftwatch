package summary_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/summary"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{Service: service, Diffs: diffs}
}

func TestCompute_Empty(t *testing.T) {
	s := summary.Compute(nil)
	if s.TotalServices != 0 || s.DriftedCount != 0 || s.CleanCount != 0 {
		t.Errorf("expected zero stats, got %+v", s)
	}
	if s.DriftRate() != 0 {
		t.Errorf("expected 0 drift rate")
	}
}

func TestCompute_AllClean(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", nil),
		makeResult("svc-b", nil),
	}
	s := summary.Compute(results)
	if s.TotalServices != 2 || s.CleanCount != 2 || s.DriftedCount != 0 {
		t.Errorf("unexpected stats: %+v", s)
	}
	if s.DriftRate() != 0 {
		t.Errorf("expected 0%% drift rate")
	}
}

func TestCompute_WithDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", []drift.Diff{{Key: "port", Expected: "8080", Actual: "9090"}}),
		makeResult("svc-b", nil),
		makeResult("svc-c", []drift.Diff{{Key: "host", Expected: "a", Actual: "b"}, {Key: "mode", Expected: "x", Actual: "y"}}),
	}
	s := summary.Compute(results)
	if s.TotalServices != 3 {
		t.Errorf("expected 3 services")
	}
	if s.DriftedCount != 2 || s.CleanCount != 1 {
		t.Errorf("unexpected drift/clean counts: %+v", s)
	}
	if s.TotalMismatches != 3 {
		t.Errorf("expected 3 total mismatches, got %d", s.TotalMismatches)
	}
	if s.ByService["svc-c"] != 2 {
		t.Errorf("expected 2 mismatches for svc-c")
	}
	rate := s.DriftRate()
	if rate < 66.6 || rate > 66.8 {
		t.Errorf("expected ~66.7%% drift rate, got %.2f", rate)
	}
}

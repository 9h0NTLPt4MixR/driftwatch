package scorecard_test

import (
	"testing"

	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/scorecard"
)

func makeResult(name string, hasDrift bool, diffs map[string][2]string) drift.Result {
	d := make([]drift.Diff, 0, len(diffs))
	for k, v := range diffs {
		d = append(d, drift.Diff{Key: k, Expected: v[0], Actual: v[1]})
	}
	return drift.Result{Service: name, HasDrift: hasDrift, Diffs: d}
}

func TestCompute_Empty(t *testing.T) {
	s := scorecard.Compute(nil)
	if s.Grade != scorecard.GradeA {
		t.Errorf("expected A, got %s", s.Grade)
	}
	if s.Pct != 100.0 {
		t.Errorf("expected 100, got %f", s.Pct)
	}
}

func TestCompute_AllClean(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", false, nil),
		makeResult("svc-b", false, nil),
	}
	s := scorecard.Compute(results)
	if s.Grade != scorecard.GradeA {
		t.Errorf("expected A, got %s", s.Grade)
	}
	if s.Pct != 100.0 {
		t.Errorf("expected 100, got %f", s.Pct)
	}
	if s.Drifted != 0 {
		t.Errorf("expected 0 drifted")
	}
}

func TestCompute_WithDrift(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", false, nil),
		makeResult("svc-b", true, map[string][2]string{"key": {"want", "got"}}),
		makeResult("svc-c", true, map[string][2]string{"x": {"1", "2"}, "y": {"a", "b"}}),
		makeResult("svc-d", false, nil),
	}
	s := scorecard.Compute(results)
	if s.Total != 4 {
		t.Errorf("expected total 4")
	}
	if s.Drifted != 2 {
		t.Errorf("expected 2 drifted")
	}
	if s.Clean != 2 {
		t.Errorf("expected 2 clean")
	}
	if s.DriftKeys != 3 {
		t.Errorf("expected 3 drift keys, got %d", s.DriftKeys)
	}
	if s.Pct != 50.0 {
		t.Errorf("expected 50.0, got %f", s.Pct)
	}
	if s.Grade != scorecard.GradeD {
		t.Errorf("expected D, got %s", s.Grade)
	}
}

func TestCompute_GradeF(t *testing.T) {
	results := []drift.Result{
		makeResult("a", true, map[string][2]string{"k": {"x", "y"}}),
		makeResult("b", true, map[string][2]string{"k": {"x", "y"}}),
		makeResult("c", true, map[string][2]string{"k": {"x", "y"}}),
	}
	s := scorecard.Compute(results)
	if s.Grade != scorecard.GradeF {
		t.Errorf("expected F, got %s", s.Grade)
	}
}

func TestScore_String(t *testing.T) {
	s := scorecard.Score{Total: 4, Clean: 3, Drifted: 1, Pct: 75.0, Grade: scorecard.GradeC, DriftKeys: 2}
	out := s.String()
	if out == "" {
		t.Error("expected non-empty string")
	}
}

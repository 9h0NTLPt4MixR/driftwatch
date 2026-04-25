package validator_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/validator"
)

func makeResult(service string, drifted bool, numDiffs int) drift.Result {
	diffs := make([]drift.Diff, numDiffs)
	for i := range diffs {
		diffs[i] = drift.Diff{Key: fmt.Sprintf("key%d", i), Expected: "a", Actual: "b"}
	}
	return drift.Result{Service: service, Drifted: drifted, Diffs: diffs}
}

func TestValidate_NoViolations(t *testing.T) {
	results := []drift.Result{
		{Service: "alpha", Drifted: false},
		{Service: "beta", Drifted: false},
	}
	rule := validator.Rule{MaxDriftedServices: 2, MaxDriftedKeys: 5}
	out := validator.Validate(results, rule)
	if !out.Passed {
		t.Fatalf("expected Passed=true, got violations: %v", out.Violations)
	}
}

func TestValidate_ExceedsMaxDriftedServices(t *testing.T) {
	results := []drift.Result{
		{Service: "alpha", Drifted: true},
		{Service: "beta", Drifted: true},
		{Service: "gamma", Drifted: false},
	}
	rule := validator.Rule{MaxDriftedServices: 1, MaxDriftedKeys: -1}
	out := validator.Validate(results, rule)
	if out.Passed {
		t.Fatal("expected Passed=false")
	}
	if len(out.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(out.Violations))
	}
	if out.Violations[0].Service != "*" {
		t.Errorf("expected service '*', got %q", out.Violations[0].Service)
	}
}

func TestValidate_ExceedsMaxDriftedKeys(t *testing.T) {
	results := []drift.Result{
		{
			Service: "svc",
			Drifted: true,
			Diffs: []drift.Diff{
				{Key: "a", Expected: "1", Actual: "2"},
				{Key: "b", Expected: "3", Actual: "4"},
				{Key: "c", Expected: "5", Actual: "6"},
			},
		},
	}
	rule := validator.Rule{MaxDriftedServices: -1, MaxDriftedKeys: 2}
	out := validator.Validate(results, rule)
	if out.Passed {
		t.Fatal("expected Passed=false")
	}
	if out.Violations[0].Service != "svc" {
		t.Errorf("expected service 'svc', got %q", out.Violations[0].Service)
	}
}

func TestValidate_RequireCleanServices_Violation(t *testing.T) {
	results := []drift.Result{
		{Service: "critical", Drifted: true, Diffs: []drift.Diff{{Key: "x", Expected: "1", Actual: "2"}}},
	}
	rule := validator.Rule{
		MaxDriftedServices:   -1,
		MaxDriftedKeys:       -1,
		RequireCleanServices: []string{"critical"},
	}
	out := validator.Validate(results, rule)
	if out.Passed {
		t.Fatal("expected Passed=false")
	}
	if out.Violations[0].Service != "critical" {
		t.Errorf("expected service 'critical', got %q", out.Violations[0].Service)
	}
}

func TestValidate_RequireCleanServices_AlreadyClean(t *testing.T) {
	results := []drift.Result{
		{Service: "critical", Drifted: false},
	}
	rule := validator.Rule{
		MaxDriftedServices:   -1,
		MaxDriftedKeys:       -1,
		RequireCleanServices: []string{"critical"},
	}
	out := validator.Validate(results, rule)
	if !out.Passed {
		t.Fatalf("expected Passed=true, got violations: %v", out.Violations)
	}
}

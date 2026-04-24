package policy_test

import (
	"testing"

	"github.com/user/driftwatch/internal/drift"
	"github.com/user/driftwatch/internal/policy"
)

func makeResult(service string, expected map[string]string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service:  service,
		Expected: expected,
		Diffs:    diffs,
	}
}

func TestEvaluate_NoViolations(t *testing.T) {
	results := []drift.Result{
		makeResult("api", map[string]string{"PORT": "8080"}, nil),
	}
	rules := []policy.Rule{{Name: "r1", MaxDriftPercent: 10}}
	got := policy.Evaluate(results, rules)
	if len(got) != 0 {
		t.Fatalf("expected no violations, got %d", len(got))
	}
}

func TestEvaluate_ProtectedKeyDrift(t *testing.T) {
	diffs := []drift.Diff{{Key: "DB_PASSWORD", Expected: "secret", Actual: "changed"}}
	results := []drift.Result{
		makeResult("api", map[string]string{"DB_PASSWORD": "secret"}, diffs),
	}
	rules := []policy.Rule{
		{Name: "no-secret-drift", ProtectedKeys: []string{"DB_PASSWORD"}, MaxDriftPercent: 100},
	}
	got := policy.Evaluate(results, rules)
	if len(got) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(got))
	}
	if got[0].Rule != "no-secret-drift" {
		t.Errorf("unexpected rule name: %s", got[0].Rule)
	}
}

func TestEvaluate_MaxDriftPercent(t *testing.T) {
	expected := map[string]string{"A": "1", "B": "2", "C": "3", "D": "4"}
	diffs := []drift.Diff{
		{Key: "A", Expected: "1", Actual: "x"},
		{Key: "B", Expected: "2", Actual: "y"},
	}
	results := []drift.Result{makeResult("svc", expected, diffs)}
	rules := []policy.Rule{{Name: "max-drift", MaxDriftPercent: 25}}
	got := policy.Evaluate(results, rules)
	if len(got) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(got))
	}
}

func TestEvaluate_ServiceFilter_Skips(t *testing.T) {
	diffs := []drift.Diff{{Key: "KEY", Expected: "a", Actual: "b"}}
	results := []drift.Result{
		makeResult("other-service", map[string]string{"KEY": "a"}, diffs),
	}
	rules := []policy.Rule{
		{Name: "api-only", Services: []string{"api"}, ProtectedKeys: []string{"KEY"}, MaxDriftPercent: 0},
	}
	got := policy.Evaluate(results, rules)
	if len(got) != 0 {
		t.Fatalf("expected no violations for unmatched service, got %d", len(got))
	}
}

func TestEvaluate_ServiceFilter_Matches(t *testing.T) {
	diffs := []drift.Diff{{Key: "KEY", Expected: "a", Actual: "b"}}
	results := []drift.Result{
		makeResult("api", map[string]string{"KEY": "a"}, diffs),
	}
	rules := []policy.Rule{
		{Name: "api-only", Services: []string{"api"}, ProtectedKeys: []string{"KEY"}, MaxDriftPercent: 100},
	}
	got := policy.Evaluate(results, rules)
	if len(got) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(got))
	}
}

func TestViolation_String(t *testing.T) {
	v := policy.Violation{Rule: "r", Service: "s", Reason: "drifted"}
	s := v.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
}

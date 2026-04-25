package labeler_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/labeler"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service: service,
		Drifted: len(diffs) > 0,
		Diffs:   diffs,
	}
}

func TestApply_NoRules(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", nil)}
	out := labeler.Apply(results, nil)
	if len(out[0].Labels) != 0 {
		t.Fatalf("expected no labels, got %v", out[0].Labels)
	}
}

func TestApply_ServiceMatch(t *testing.T) {
	rules := []labeler.Rule{{Label: "payments", Service: "pay-svc"}}
	results := []drift.Result{
		makeResult("pay-svc", nil),
		makeResult("auth-svc", nil),
	}
	out := labeler.Apply(results, rules)
	if len(out[0].Labels) != 1 || out[0].Labels[0] != "payments" {
		t.Fatalf("expected label 'payments' on pay-svc, got %v", out[0].Labels)
	}
	if len(out[1].Labels) != 0 {
		t.Fatalf("expected no labels on auth-svc, got %v", out[1].Labels)
	}
}

func TestApply_KeyPrefixMatch(t *testing.T) {
	rules := []labeler.Rule{{Label: "auth-critical", KeyPrefix: "auth."}}
	diffs := []drift.Diff{{Key: "auth.token_ttl", Expected: "3600", Actual: "7200"}}
	results := []drift.Result{makeResult("svc-a", diffs)}
	out := labeler.Apply(results, rules)
	if len(out[0].Labels) != 1 || out[0].Labels[0] != "auth-critical" {
		t.Fatalf("expected 'auth-critical' label, got %v", out[0].Labels)
	}
}

func TestApply_MinDrifted_NotMet(t *testing.T) {
	rules := []labeler.Rule{{Label: "heavy-drift", MinDrifted: 3}}
	diffs := []drift.Diff{{Key: "k1", Expected: "a", Actual: "b"}}
	results := []drift.Result{makeResult("svc-a", diffs)}
	out := labeler.Apply(results, rules)
	if len(out[0].Labels) != 0 {
		t.Fatalf("expected no labels, got %v", out[0].Labels)
	}
}

func TestApply_MinDrifted_Met(t *testing.T) {
	rules := []labeler.Rule{{Label: "heavy-drift", MinDrifted: 2}}
	diffs := []drift.Diff{
		{Key: "k1", Expected: "a", Actual: "b"},
		{Key: "k2", Expected: "x", Actual: "y"},
	}
	results := []drift.Result{makeResult("svc-a", diffs)}
	out := labeler.Apply(results, rules)
	if len(out[0].Labels) != 1 || out[0].Labels[0] != "heavy-drift" {
		t.Fatalf("expected 'heavy-drift' label, got %v", out[0].Labels)
	}
}

func TestApply_DeduplicatesLabels(t *testing.T) {
	rules := []labeler.Rule{
		{Label: "important", Service: "svc-a"},
		{Label: "important", MinDrifted: 0},
	}
	results := []drift.Result{makeResult("svc-a", nil)}
	out := labeler.Apply(results, rules)
	if len(out[0].Labels) != 1 {
		t.Fatalf("expected exactly 1 label, got %v", out[0].Labels)
	}
}

func TestFilterByLabel_ReturnsMatches(t *testing.T) {
	r1 := makeResult("svc-a", nil)
	r1.Labels = []string{"critical"}
	r2 := makeResult("svc-b", nil)
	r2.Labels = []string{"low-priority"}
	out := labeler.FilterByLabel([]drift.Result{r1, r2}, "critical")
	if len(out) != 1 || out[0].Service != "svc-a" {
		t.Fatalf("expected svc-a only, got %v", out)
	}
}

func TestFilterByLabel_NoMatches(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", nil)}
	out := labeler.FilterByLabel(results, "nonexistent")
	if len(out) != 0 {
		t.Fatalf("expected empty, got %v", out)
	}
}

package grouper_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/grouper"
)

func makeResult(service string, hasDrift bool, tags map[string]string) drift.Result {
	return drift.Result{
		Service:  service,
		HasDrift: hasDrift,
		Tags:     tags,
	}
}

func TestCompute_GroupByService(t *testing.T) {
	results := []drift.Result{
		makeResult("api", true, nil),
		makeResult("worker", false, nil),
		makeResult("api", false, nil),
	}
	groups := grouper.Compute(results, grouper.Options{By: grouper.GroupByService})
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "api" || len(groups[0].Results) != 2 {
		t.Errorf("unexpected api group: %+v", groups[0])
	}
}

func TestCompute_GroupByStatus(t *testing.T) {
	results := []drift.Result{
		makeResult("api", true, nil),
		makeResult("worker", false, nil),
		makeResult("db", true, nil),
	}
	groups := grouper.Compute(results, grouper.Options{By: grouper.GroupByStatus})
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	for _, g := range groups {
		if g.Name == "drifted" && len(g.Results) != 2 {
			t.Errorf("expected 2 drifted, got %d", len(g.Results))
		}
		if g.Name == "clean" && len(g.Results) != 1 {
			t.Errorf("expected 1 clean, got %d", len(g.Results))
		}
	}
}

func TestCompute_GroupByTag_WithUntagged(t *testing.T) {
	results := []drift.Result{
		makeResult("api", true, map[string]string{"env": "prod"}),
		makeResult("worker", false, map[string]string{"env": "staging"}),
		makeResult("db", true, nil),
	}
	groups := grouper.Compute(results, grouper.Options{By: grouper.GroupByTag, TagKey: "env"})
	names := map[string]int{}
	for _, g := range groups {
		names[g.Name] = len(g.Results)
	}
	if names["prod"] != 1 || names["staging"] != 1 || names["(untagged)"] != 1 {
		t.Errorf("unexpected groups: %v", names)
	}
}

func TestCompute_GroupByTag_MissingTagKey(t *testing.T) {
	// When GroupByTag is used but TagKey is empty, all results should fall
	// into the "(untagged)" bucket since no tag key can be matched.
	results := []drift.Result{
		makeResult("api", true, map[string]string{"env": "prod"}),
		makeResult("worker", false, map[string]string{"env": "staging"}),
	}
	groups := grouper.Compute(results, grouper.Options{By: grouper.GroupByTag, TagKey: ""})
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Name != "(untagged)" || len(groups[0].Results) != 2 {
		t.Errorf("unexpected group: %+v", groups[0])
	}
}

func TestCompute_Empty(t *testing.T) {
	groups := grouper.Compute(nil, grouper.Options{By: grouper.GroupByService})
	if len(groups) != 0 {
		t.Errorf("expected empty groups, got %d", len(groups))
	}
}

func TestSummary(t *testing.T) {
	results := []drift.Result{
		makeResult("api", true, nil),
		makeResult("api", false, nil),
	}
	groups := grouper.Compute(results, grouper.Options{By: grouper.GroupByService})
	lines := grouper.Summary(groups)
	if len(lines) != 1 {
		t.Fatalf("expected 1 summary line, got %d", len(lines))
	}
	if lines[0] != "api: 1/2 drifted" {
		t.Errorf("unexpected summary: %q", lines[0])
	}
}

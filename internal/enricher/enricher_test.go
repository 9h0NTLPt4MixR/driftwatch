package enricher_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/enricher"
)

func makeResult(service string, drifted bool) drift.Result {
	r := drift.Result{Service: service}
	if drifted {
		r.Diffs = []drift.Diff{{Key: "replicas", Expected: "3", Actual: "1"}}
	}
	return r
}

func TestApply_NoMeta(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", false)}
	enriched := enricher.Apply(results, enricher.Options{})
	if len(enriched) != 1 {
		t.Fatalf("expected 1 result, got %d", len(enriched))
	}
	if enriched[0].Meta.Environment != "" {
		t.Errorf("expected empty environment, got %q", enriched[0].Meta.Environment)
	}
}

func TestApply_DefaultEnvironment(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", false)}
	opts := enricher.Options{DefaultEnvironment: "staging"}
	enriched := enricher.Apply(results, opts)
	if enriched[0].Meta.Environment != "staging" {
		t.Errorf("expected staging, got %q", enriched[0].Meta.Environment)
	}
}

func TestApply_PerServiceMeta(t *testing.T) {
	results := []drift.Result{makeResult("api", true), makeResult("worker", false)}
	opts := enricher.Options{
		DefaultEnvironment: "production",
		ServiceMeta: map[string]enricher.Meta{
			"api": {Owner: "platform", Tier: "critical", Environment: "production"},
		},
	}
	enriched := enricher.Apply(results, opts)
	if enriched[0].Meta.Owner != "platform" {
		t.Errorf("expected owner platform, got %q", enriched[0].Meta.Owner)
	}
	if enriched[1].Meta.Owner != "" {
		t.Errorf("expected empty owner for worker, got %q", enriched[1].Meta.Owner)
	}
}

func TestApply_PerServiceMeta_InheritsDefault(t *testing.T) {
	results := []drift.Result{makeResult("svc", false)}
	opts := enricher.Options{
		DefaultEnvironment: "prod",
		ServiceMeta: map[string]enricher.Meta{
			"svc": {Owner: "team-a"}, // no Environment set
		},
	}
	enriched := enricher.Apply(results, opts)
	if enriched[0].Meta.Environment != "prod" {
		t.Errorf("expected prod from default, got %q", enriched[0].Meta.Environment)
	}
}

func TestFilterByEnvironment(t *testing.T) {
	results := []enricher.Result{
		{Result: makeResult("a", false), Meta: enricher.Meta{Environment: "prod"}},
		{Result: makeResult("b", true), Meta: enricher.Meta{Environment: "staging"}},
		{Result: makeResult("c", false), Meta: enricher.Meta{Environment: "prod"}},
	}
	filtered := enricher.FilterByEnvironment(results, "prod")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 prod results, got %d", len(filtered))
	}
}

func TestFilterByEnvironment_Empty(t *testing.T) {
	results := []enricher.Result{
		{Result: makeResult("a", false), Meta: enricher.Meta{Environment: "prod"}},
	}
	filtered := enricher.FilterByEnvironment(results, "")
	if len(filtered) != 1 {
		t.Fatalf("expected all results when env empty, got %d", len(filtered))
	}
}

func TestFilterByOwner(t *testing.T) {
	results := []enricher.Result{
		{Result: makeResult("a", false), Meta: enricher.Meta{Owner: "platform"}},
		{Result: makeResult("b", true), Meta: enricher.Meta{Owner: "data"}},
	}
	filtered := enricher.FilterByOwner(results, "Platform") // case-insensitive
	if len(filtered) != 1 {
		t.Fatalf("expected 1 result, got %d", len(filtered))
	}
	if filtered[0].Service != "a" {
		t.Errorf("expected service a, got %q", filtered[0].Service)
	}
}

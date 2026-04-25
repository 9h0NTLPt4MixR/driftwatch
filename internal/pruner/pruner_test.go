package pruner_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/pruner"
)

func makeResult(service string, hasDrift bool, age time.Duration) drift.Result {
	return drift.Result{
		Service:   service,
		HasDrift:  hasDrift,
		CheckedAt: time.Now().Add(-age),
	}
}

func TestPrune_NoOptions_ReturnsAll(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-b", false, 0),
	}
	got := pruner.Prune(results, pruner.Options{})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestPrune_RemoveClean(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-b", false, 0),
		makeResult("svc-c", false, 0),
	}
	got := pruner.Prune(results, pruner.Options{RemoveClean: true})
	if len(got) != 1 || got[0].Service != "svc-a" {
		t.Fatalf("expected only svc-a, got %+v", got)
	}
}

func TestPrune_ExcludeServices(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-b", true, 0),
		makeResult("svc-c", true, 0),
	}
	got := pruner.Prune(results, pruner.Options{ExcludeServices: []string{"svc-b", "svc-c"}})
	if len(got) != 1 || got[0].Service != "svc-a" {
		t.Fatalf("unexpected results: %+v", got)
	}
}

func TestPrune_MaxAge_RemovesOld(t *testing.T) {
	results := []drift.Result{
		makeResult("fresh", true, 1*time.Minute),
		makeResult("stale", true, 2*time.Hour),
	}
	got := pruner.Prune(results, pruner.Options{MaxAge: 30 * time.Minute})
	if len(got) != 1 || got[0].Service != "fresh" {
		t.Fatalf("expected only fresh, got %+v", got)
	}
}

func TestPrune_ZeroCheckedAt_Retained(t *testing.T) {
	r := drift.Result{Service: "no-ts", HasDrift: true} // zero CheckedAt
	got := pruner.Prune([]drift.Result{r}, pruner.Options{MaxAge: 1 * time.Minute})
	if len(got) != 1 {
		t.Fatalf("result with zero CheckedAt should be retained, got %d", len(got))
	}
}

func TestPruneWithStats(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-b", false, 0),
		makeResult("svc-c", false, 0),
	}
	_, stats := pruner.PruneWithStats(results, pruner.Options{RemoveClean: true})
	if stats.Total != 3 || stats.Retained != 1 || stats.Removed != 2 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

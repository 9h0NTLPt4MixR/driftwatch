package indexer_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/indexer"
)

func makeResult(service string, drifted bool, keys ...string) drift.Result {
	r := drift.Result{
		Service:   service,
		Drifted:   drifted,
		Timestamp: time.Now(),
	}
	for _, k := range keys {
		r.Diffs = append(r.Diffs, drift.Diff{Key: k, Expected: "a", Actual: "b"})
	}
	return r
}

func TestBuild_NilResults(t *testing.T) {
	_, err := indexer.Build(nil)
	if err == nil {
		t.Fatal("expected error for nil results")
	}
}

func TestBuild_Empty(t *testing.T) {
	idx, err := indexer.Build([]drift.Result{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(idx.ByService) != 0 {
		t.Errorf("expected empty ByService")
	}
}

func TestLookupService_Found(t *testing.T) {
	results := []drift.Result{makeResult("auth", true, "timeout")}
	idx, _ := indexer.Build(results)
	r, ok := idx.LookupService("AUTH")
	if !ok {
		t.Fatal("expected to find service")
	}
	if r.Service != "auth" {
		t.Errorf("unexpected service: %s", r.Service)
	}
}

func TestLookupService_NotFound(t *testing.T) {
	idx, _ := indexer.Build([]drift.Result{})
	_, ok := idx.LookupService("ghost")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestDriftedAndClean_Partitioned(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, "key1"),
		makeResult("svc-b", false),
		makeResult("svc-c", true, "key2"),
	}
	idx, _ := indexer.Build(results)
	if len(idx.Drifted) != 2 {
		t.Errorf("expected 2 drifted, got %d", len(idx.Drifted))
	}
	if len(idx.Clean) != 1 {
		t.Errorf("expected 1 clean, got %d", len(idx.Clean))
	}
}

func TestLookupKey_ReturnsMatchingServices(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, "timeout", "retries"),
		makeResult("svc-b", true, "timeout"),
		makeResult("svc-c", true, "retries"),
	}
	idx, _ := indexer.Build(results)
	hits := idx.LookupKey("timeout")
	if len(hits) != 2 {
		t.Errorf("expected 2 services with 'timeout', got %d", len(hits))
	}
}

func TestStats_Format(t *testing.T) {
	results := []drift.Result{makeResult("svc", true, "k")}
	idx, _ := indexer.Build(results)
	s := idx.Stats()
	if s == "" {
		t.Error("expected non-empty stats string")
	}
}

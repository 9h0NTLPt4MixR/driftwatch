package trimmer_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/trimmer"
)

func makeResult(service string, drifted bool, age time.Duration) drift.Result {
	return drift.Result{
		Service:   service,
		Drifted:   drifted,
		Timestamp: time.Now().Add(-age),
	}
}

func TestTrim_NoOptions_ReturnsAll(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-b", false, 0),
	}
	out := trimmer.Trim(results, trimmer.Options{})
	if len(out.Kept) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(out.Kept))
	}
	if out.Removed != 0 {
		t.Fatalf("expected 0 removed, got %d", out.Removed)
	}
}

func TestTrim_OnlyDrifted_DropsClean(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-b", false, 0),
		makeResult("svc-c", true, 0),
	}
	out := trimmer.Trim(results, trimmer.Options{OnlyDrifted: true})
	if len(out.Kept) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(out.Kept))
	}
	if out.Removed != 1 {
		t.Fatalf("expected 1 removed, got %d", out.Removed)
	}
}

func TestTrim_MaxAge_RemovesOld(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 10*time.Minute),
		makeResult("svc-b", true, 2*time.Hour),
		makeResult("svc-c", false, 30*time.Second),
	}
	out := trimmer.Trim(results, trimmer.Options{MaxAge: 1 * time.Hour})
	if len(out.Kept) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(out.Kept))
	}
	if out.Removed != 1 {
		t.Fatalf("expected 1 removed, got %d", out.Removed)
	}
}

func TestTrim_MaxPerService_LimitsEntries(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-a", true, 5*time.Minute),
		makeResult("svc-a", true, 10*time.Minute),
		makeResult("svc-b", false, 0),
	}
	out := trimmer.Trim(results, trimmer.Options{MaxPerService: 2})
	if len(out.Kept) != 3 {
		t.Fatalf("expected 3 kept, got %d", len(out.Kept))
	}
	if out.Removed != 1 {
		t.Fatalf("expected 1 removed, got %d", out.Removed)
	}
}

func TestTrim_CombinedOptions(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 0),
		makeResult("svc-a", true, 0),
		makeResult("svc-a", true, 0),
		makeResult("svc-b", false, 0),
		makeResult("svc-c", true, 3*time.Hour),
	}
	out := trimmer.Trim(results, trimmer.Options{
		OnlyDrifted:   true,
		MaxPerService: 2,
		MaxAge:        2 * time.Hour,
	})
	// svc-b dropped (clean), svc-c dropped (too old), svc-a capped at 2
	if len(out.Kept) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(out.Kept))
	}
	if out.Removed != 3 {
		t.Fatalf("expected 3 removed, got %d", out.Removed)
	}
}

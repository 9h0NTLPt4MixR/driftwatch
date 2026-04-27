package retention_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/retention"
)

func makeResult(service string, hasDrift bool, age time.Duration) drift.Result {
	return drift.Result{
		Service:   service,
		HasDrift:  hasDrift,
		Timestamp: time.Now().Add(-age),
	}
}

func TestApply_NoOptions_ReturnsAll(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", false, time.Hour),
		makeResult("svc-b", true, 48*time.Hour),
	}
	out, err := retention.Apply(results, retention.Policy{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Retained) != 2 {
		t.Errorf("expected 2 retained, got %d", len(out.Retained))
	}
	if out.Removed != 0 {
		t.Errorf("expected 0 removed, got %d", out.Removed)
	}
}

func TestApply_MaxAge_RemovesOld(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", false, 2*time.Hour),
		makeResult("svc-b", false, 10*time.Minute),
	}
	out, err := retention.Apply(results, retention.Policy{MaxAge: time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Retained) != 1 {
		t.Errorf("expected 1 retained, got %d", len(out.Retained))
	}
	if out.Retained[0].Service != "svc-b" {
		t.Errorf("expected svc-b retained, got %s", out.Retained[0].Service)
	}
	if out.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", out.Removed)
	}
}

func TestApply_KeepDrifted_ExemptsFromAge(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", true, 72*time.Hour),
		makeResult("svc-b", false, 72*time.Hour),
	}
	out, err := retention.Apply(results, retention.Policy{
		MaxAge:      time.Hour,
		KeepDrifted: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Retained) != 1 || out.Retained[0].Service != "svc-a" {
		t.Errorf("expected only svc-a retained, got %+v", out.Retained)
	}
}

func TestApply_MaxEntries_KeepsNewest(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", false, 3*time.Hour),
		makeResult("svc-b", false, 2*time.Hour),
		makeResult("svc-c", false, time.Hour),
	}
	out, err := retention.Apply(results, retention.Policy{MaxEntries: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Retained) != 2 {
		t.Errorf("expected 2 retained, got %d", len(out.Retained))
	}
	if out.Retained[0].Service != "svc-b" || out.Retained[1].Service != "svc-c" {
		t.Errorf("unexpected retained order: %+v", out.Retained)
	}
}

func TestApply_InvalidMaxAge_ReturnsError(t *testing.T) {
	_, err := retention.Apply(nil, retention.Policy{MaxAge: -time.Second})
	if err == nil {
		t.Error("expected error for negative MaxAge")
	}
}

func TestApply_InvalidMaxEntries_ReturnsError(t *testing.T) {
	_, err := retention.Apply(nil, retention.Policy{MaxEntries: -1})
	if err == nil {
		t.Error("expected error for negative MaxEntries")
	}
}

func TestStats_ReturnsCounts(t *testing.T) {
	r := retention.Result{Retained: []drift.Result{{}, {}}, Removed: 3}
	s := retention.Stats(r)
	if s["retained"] != 2 || s["removed"] != 3 {
		t.Errorf("unexpected stats: %v", s)
	}
}

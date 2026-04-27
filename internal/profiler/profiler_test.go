package profiler_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/profiler"
)

func TestCompute_NoEntries_ReturnsError(t *testing.T) {
	p := profiler.New()
	_, err := p.Compute()
	if err == nil {
		t.Fatal("expected error for empty profiler, got nil")
	}
}

func TestRecord_And_Compute_SingleEntry(t *testing.T) {
	p := profiler.New()
	p.Record("svc-a", 50*time.Millisecond, false)

	stats, err := p.Compute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 1 {
		t.Errorf("expected Total=1, got %d", stats.Total)
	}
	if stats.Min != 50*time.Millisecond {
		t.Errorf("expected Min=50ms, got %v", stats.Min)
	}
	if stats.Max != 50*time.Millisecond {
		t.Errorf("expected Max=50ms, got %v", stats.Max)
	}
	if stats.Avg != 50*time.Millisecond {
		t.Errorf("expected Avg=50ms, got %v", stats.Avg)
	}
}

func TestCompute_MultipleEntries_Aggregates(t *testing.T) {
	p := profiler.New()
	p.Record("svc-a", 10*time.Millisecond, false)
	p.Record("svc-b", 30*time.Millisecond, true)
	p.Record("svc-c", 20*time.Millisecond, false)

	stats, err := p.Compute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 3 {
		t.Errorf("expected Total=3, got %d", stats.Total)
	}
	if stats.Min != 10*time.Millisecond {
		t.Errorf("expected Min=10ms, got %v", stats.Min)
	}
	if stats.Max != 30*time.Millisecond {
		t.Errorf("expected Max=30ms, got %v", stats.Max)
	}
	if stats.Avg != 20*time.Millisecond {
		t.Errorf("expected Avg=20ms, got %v", stats.Avg)
	}
}

func TestCompute_EntriesSortedSlowestFirst(t *testing.T) {
	p := profiler.New()
	p.Record("fast", 5*time.Millisecond, false)
	p.Record("slow", 100*time.Millisecond, true)
	p.Record("mid", 40*time.Millisecond, false)

	stats, err := p.Compute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Entries[0].Service != "slow" {
		t.Errorf("expected first entry to be 'slow', got %q", stats.Entries[0].Service)
	}
	if stats.Entries[2].Service != "fast" {
		t.Errorf("expected last entry to be 'fast', got %q", stats.Entries[2].Service)
	}
}

func TestRecord_DriftedFlag_Preserved(t *testing.T) {
	p := profiler.New()
	p.Record("svc-x", 15*time.Millisecond, true)

	stats, _ := p.Compute()
	if !stats.Entries[0].Drifted {
		t.Error("expected Drifted=true for svc-x")
	}
}

package rollup_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/rollup"
)

func makeResult(ts time.Time, hasDrift bool, diffs int) drift.Result {
	d := make([]drift.Diff, diffs)
	return drift.Result{
		Service:   "svc",
		HasDrift:  hasDrift,
		Diffs:     d,
		Timestamp: ts,
	}
}

func TestCompute_Empty(t *testing.T) {
	buckets := rollup.Compute(nil, rollup.Daily)
	if len(buckets) != 0 {
		t.Fatalf("expected 0 buckets, got %d", len(buckets))
	}
}

func TestCompute_Daily_SingleBucket(t *testing.T) {
	base := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	results := []drift.Result{
		makeResult(base, true, 3),
		makeResult(base.Add(2*time.Hour), false, 0),
		makeResult(base.Add(5*time.Hour), true, 1),
	}
	buckets := rollup.Compute(results, rollup.Daily)
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	b := buckets[0]
	if b.TotalServices != 3 {
		t.Errorf("TotalServices: want 3, got %d", b.TotalServices)
	}
	if b.DriftedCount != 2 {
		t.Errorf("DriftedCount: want 2, got %d", b.DriftedCount)
	}
	if b.CleanCount != 1 {
		t.Errorf("CleanCount: want 1, got %d", b.CleanCount)
	}
	if b.TotalDiffs != 4 {
		t.Errorf("TotalDiffs: want 4, got %d", b.TotalDiffs)
	}
}

func TestCompute_Daily_MultipleBuckets_Sorted(t *testing.T) {
	day1 := time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC)
	day2 := time.Date(2024, 6, 3, 8, 0, 0, 0, time.UTC)
	results := []drift.Result{
		makeResult(day2, true, 2),
		makeResult(day1, false, 0),
	}
	buckets := rollup.Compute(results, rollup.Daily)
	if len(buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	if !buckets[0].WindowStart.Before(buckets[1].WindowStart) {
		t.Error("expected buckets sorted chronologically")
	}
}

func TestCompute_Hourly_BucketKey(t *testing.T) {
	base := time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC)
	results := []drift.Result{
		makeResult(base.Add(10*time.Minute), true, 1),
		makeResult(base.Add(45*time.Minute), true, 2),
		makeResult(base.Add(90*time.Minute), false, 0),
	}
	buckets := rollup.Compute(results, rollup.Hourly)
	if len(buckets) != 2 {
		t.Fatalf("expected 2 hourly buckets, got %d", len(buckets))
	}
}

func TestCompute_DriftRate(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	results := []drift.Result{
		makeResult(base, true, 1),
		makeResult(base.Add(time.Hour), true, 1),
		makeResult(base.Add(2*time.Hour), false, 0),
		makeResult(base.Add(3*time.Hour), false, 0),
	}
	buckets := rollup.Compute(results, rollup.Daily)
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket")
	}
	if buckets[0].DriftRate != 0.5 {
		t.Errorf("DriftRate: want 0.5, got %f", buckets[0].DriftRate)
	}
}

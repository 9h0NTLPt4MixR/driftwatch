package sorter_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/sorter"
)

func makeResult(service string, diffs int, scannedAt string) drift.Result {
	d := make([]drift.Diff, diffs)
	for i := range d {
		d[i] = drift.Diff{Key: "k", Expected: "a", Actual: "b"}
	}
	return drift.Result{Service: service, Diffs: d, ScannedAt: scannedAt}
}

func TestSort_ByService_Ascending(t *testing.T) {
	input := []drift.Result{
		makeResult("zebra", 0, ""),
		makeResult("alpha", 0, ""),
		makeResult("mango", 0, ""),
	}
	out := sorter.Sort(input, sorter.Options{By: sorter.ByService})
	if out[0].Service != "alpha" || out[1].Service != "mango" || out[2].Service != "zebra" {
		t.Errorf("unexpected order: %v", serviceNames(out))
	}
}

func TestSort_ByService_Descending(t *testing.T) {
	input := []drift.Result{
		makeResult("alpha", 0, ""),
		makeResult("zebra", 0, ""),
	}
	out := sorter.Sort(input, sorter.Options{By: sorter.ByService, Descending: true})
	if out[0].Service != "zebra" {
		t.Errorf("expected zebra first, got %s", out[0].Service)
	}
}

func TestSort_ByDriftCount_Ascending(t *testing.T) {
	input := []drift.Result{
		makeResult("svc-c", 5, ""),
		makeResult("svc-a", 1, ""),
		makeResult("svc-b", 3, ""),
	}
	out := sorter.Sort(input, sorter.Options{By: sorter.ByDriftCount})
	if len(out[0].Diffs) != 1 || len(out[2].Diffs) != 5 {
		t.Errorf("unexpected drift order: %v", driftCounts(out))
	}
}

func TestSort_ByScannedAt_Ascending(t *testing.T) {
	now := time.Now().UTC()
	input := []drift.Result{
		makeResult("svc-b", 0, now.Add(2*time.Hour).Format(time.RFC3339)),
		makeResult("svc-a", 0, now.Format(time.RFC3339)),
		makeResult("svc-c", 0, now.Add(time.Hour).Format(time.RFC3339)),
	}
	out := sorter.Sort(input, sorter.Options{By: sorter.ByScannedAt})
	if out[0].Service != "svc-a" || out[2].Service != "svc-b" {
		t.Errorf("unexpected time order: %v", serviceNames(out))
	}
}

func TestSort_DoesNotMutateInput(t *testing.T) {
	input := []drift.Result{
		makeResult("z", 0, ""),
		makeResult("a", 0, ""),
	}
	_ = sorter.Sort(input, sorter.Options{By: sorter.ByService})
	if input[0].Service != "z" {
		t.Error("Sort mutated the original slice")
	}
}

func serviceNames(results []drift.Result) []string {
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.Service
	}
	return names
}

func driftCounts(results []drift.Result) []int {
	counts := make([]int, len(results))
	for i, r := range results {
		counts[i] = len(r.Diffs)
	}
	return counts
}

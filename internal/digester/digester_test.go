package digester_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/digester"
	"github.com/driftwatch/internal/drift"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service:   service,
		HasDrift:  len(diffs) > 0,
		Diffs:     diffs,
		FetchedAt: time.Now(),
	}
}

func TestCompute_Empty(t *testing.T) {
	out, err := digester.Compute(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty slice, got %d", len(out))
	}
}

func TestCompute_NoDrift(t *testing.T) {
	r := makeResult("api", nil)
	out, err := digester.Compute([]drift.Result{r})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Service != "api" {
		t.Errorf("expected service api, got %s", out[0].Service)
	}
	if out[0].Drifted {
		t.Error("expected Drifted=false")
	}
	if out[0].Digest == "" {
		t.Error("expected non-empty digest")
	}
}

func TestCompute_Stable_DiffOrder(t *testing.T) {
	diffs1 := []drift.Diff{
		{Key: "a", Expected: "1", Actual: "2"},
		{Key: "b", Expected: "x", Actual: "y"},
	}
	diffs2 := []drift.Diff{
		{Key: "b", Expected: "x", Actual: "y"},
		{Key: "a", Expected: "1", Actual: "2"},
	}
	out1, _ := digester.Compute([]drift.Result{makeResult("svc", diffs1)})
	out2, _ := digester.Compute([]drift.Result{makeResult("svc", diffs2)})
	if out1[0].Digest != out2[0].Digest {
		t.Errorf("digests differ for same diffs in different order: %s vs %s",
			out1[0].Digest, out2[0].Digest)
	}
}

func TestChanged_DetectsNewDrift(t *testing.T) {
	prev, _ := digester.Compute([]drift.Result{makeResult("api", nil)})
	next, _ := digester.Compute([]drift.Result{
		makeResult("api", []drift.Diff{{Key: "port", Expected: "8080", Actual: "9090"}}),
	})
	changed := digester.Changed(prev, next)
	if len(changed) != 1 || changed[0] != "api" {
		t.Errorf("expected [api], got %v", changed)
	}
}

func TestChanged_NoChange(t *testing.T) {
	results := []drift.Result{makeResult("api", nil)}
	prev, _ := digester.Compute(results)
	next, _ := digester.Compute(results)
	if got := digester.Changed(prev, next); len(got) != 0 {
		t.Errorf("expected no changes, got %v", got)
	}
}

func TestChanged_NewService(t *testing.T) {
	prev, _ := digester.Compute([]drift.Result{makeResult("api", nil)})
	next, _ := digester.Compute([]drift.Result{
		makeResult("api", nil),
		makeResult("worker", nil),
	})
	changed := digester.Changed(prev, next)
	if len(changed) != 1 || changed[0] != "worker" {
		t.Errorf("expected [worker], got %v", changed)
	}
}

package patcher_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/patcher"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service: service,
		Drifted: len(diffs) > 0,
		Diffs:   diffs,
	}
}

func diff(key, expected, actual string) drift.Diff {
	return drift.Diff{Key: key, Expected: expected, Actual: actual, Drifted: expected != actual}
}

func TestApply_NoOps_Unchanged(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", []drift.Diff{diff("port", "8080", "9090")})}
	out := patcher.Apply(results, nil)
	if len(out) != 1 || len(out[0].Diffs) != 1 {
		t.Fatal("expected unchanged result")
	}
}

func TestApply_Suppress_RemovesDiff(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", []drift.Diff{diff("port", "8080", "9090")})}
	ops := []patcher.Op{{Service: "svc-a", Key: "port", Action: "suppress"}}
	out := patcher.Apply(results, ops)
	if len(out[0].Diffs) != 0 {
		t.Errorf("expected diff suppressed, got %d diffs", len(out[0].Diffs))
	}
	if out[0].Drifted {
		t.Error("expected Drifted=false after suppression")
	}
	if len(out[0].PatchedKeys) != 1 || out[0].PatchedKeys[0] != "port" {
		t.Errorf("expected PatchedKeys=[port], got %v", out[0].PatchedKeys)
	}
}

func TestApply_Override_UpdatesActual(t *testing.T) {
	results := []drift.Result{makeResult("svc-b", []drift.Diff{diff("log_level", "info", "debug")})}
	ops := []patcher.Op{{Service: "svc-b", Key: "log_level", Action: "override", Value: "info"}}
	out := patcher.Apply(results, ops)
	if len(out[0].Diffs) != 1 {
		t.Fatal("expected diff retained after override")
	}
	if out[0].Diffs[0].Actual != "info" {
		t.Errorf("expected Actual=info, got %s", out[0].Diffs[0].Actual)
	}
	if out[0].Diffs[0].Drifted {
		t.Error("expected Drifted=false after matching override")
	}
}

func TestApply_ServiceFilter_SkipsOtherServices(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", []drift.Diff{diff("port", "8080", "9090")}),
		makeResult("svc-b", []drift.Diff{diff("port", "8080", "9090")}),
	}
	ops := []patcher.Op{{Service: "svc-a", Key: "port", Action: "suppress"}}
	out := patcher.Apply(results, ops)
	if len(out[0].Diffs) != 0 {
		t.Error("svc-a: expected diff suppressed")
	}
	if len(out[1].Diffs) != 1 {
		t.Error("svc-b: expected diff retained")
	}
}

func TestApply_GlobalOp_AffectsAllServices(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", []drift.Diff{diff("debug", "false", "true")}),
		makeResult("svc-b", []drift.Diff{diff("debug", "false", "true")}),
	}
	ops := []patcher.Op{{Key: "debug", Action: "suppress"}}
	out := patcher.Apply(results, ops)
	for _, r := range out {
		if len(r.Diffs) != 0 {
			t.Errorf("%s: expected diff suppressed by global op", r.Service)
		}
	}
}

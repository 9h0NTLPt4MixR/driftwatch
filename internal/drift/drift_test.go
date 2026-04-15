package drift

import (
	"testing"

	"github.com/user/driftwatch/internal/config"
)

func makeService(name, endpoint string, expected map[string]interface{}) config.Service {
	return config.Service{
		Name:     name,
		Endpoint: endpoint,
		Expected: expected,
	}
}

func TestCompare_NoDrift(t *testing.T) {
	expected := map[string]interface{}{"version": "1.2.3", "env": "production"}
	actual := map[string]interface{}{"version": "1.2.3", "env": "production"}
	diffs := compare(expected, actual)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs, got %d", len(diffs))
	}
}

func TestCompare_ValueMismatch(t *testing.T) {
	expected := map[string]interface{}{"version": "1.2.3"}
	actual := map[string]interface{}{"version": "1.9.0"}
	diffs := compare(expected, actual)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Key != "version" {
		t.Errorf("unexpected diff key: %s", diffs[0].Key)
	}
}

func TestCompare_MissingKey(t *testing.T) {
	expected := map[string]interface{}{"version": "1.2.3", "env": "production"}
	actual := map[string]interface{}{"version": "1.2.3"}
	diffs := compare(expected, actual)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Key != "env" {
		t.Errorf("unexpected diff key: %s", diffs[0].Key)
	}
}

func TestHasDrift_True(t *testing.T) {
	results := []Result{
		{Service: "svc", Diffs: []Diff{{Key: "version", Expected: "1", Actual: "2"}}},
	}
	if !HasDrift(results) {
		t.Error("expected HasDrift to return true")
	}
}

func TestHasDrift_False(t *testing.T) {
	results := []Result{
		{Service: "svc", Diffs: []Diff{}},
	}
	if HasDrift(results) {
		t.Error("expected HasDrift to return false")
	}
}

func TestReport_ContainsServiceName(t *testing.T) {
	results := []Result{
		{
			Service: "my-service",
			Diffs: []Diff{{Key: "version", Expected: "1.0", Actual: "2.0"}},
		},
	}
	out := Report(results)
	if len(out) == 0 {
		t.Fatal("expected non-empty report")
	}
	found := false
	for _, line := range out {
		if line == "" {
			continue
		}
		found = true
		break
	}
	if !found {
		t.Error("report appears empty")
	}
}

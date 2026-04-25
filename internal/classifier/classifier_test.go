package classifier_test

import (
	"testing"

	"github.com/driftwatch/internal/classifier"
	"github.com/driftwatch/internal/drift"
)

func makeResult(service string, keys ...string) drift.Result {
	diffs := make([]drift.Diff, len(keys))
	for i, k := range keys {
		diffs[i] = drift.Diff{Key: k, Expected: "a", Actual: "b"}
	}
	return drift.Result{
		Service:  service,
		HasDrift: len(keys) > 0,
		Diffs:    diffs,
	}
}

func TestClassify_NoDrift(t *testing.T) {
	results := []drift.Result{makeResult("svc-a")}
	out := classifier.Classify(results, classifier.Options{})
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Level != classifier.LevelNone {
		t.Errorf("expected none, got %s", out[0].Level)
	}
}

func TestClassify_LowDrift(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", "key1")}
	out := classifier.Classify(results, classifier.Options{})
	if out[0].Level != classifier.LevelLow {
		t.Errorf("expected low, got %s", out[0].Level)
	}
}

func TestClassify_MediumDrift(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", "k1", "k2", "k3")}
	out := classifier.Classify(results, classifier.Options{})
	if out[0].Level != classifier.LevelMedium {
		t.Errorf("expected medium, got %s", out[0].Level)
	}
}

func TestClassify_HighDrift(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", "k1", "k2", "k3", "k4", "k5", "k6")}
	out := classifier.Classify(results, classifier.Options{})
	if out[0].Level != classifier.LevelHigh {
		t.Errorf("expected high, got %s", out[0].Level)
	}
}

func TestClassify_CriticalOnProtectedKey(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", "DB_PASSWORD")}
	opts := classifier.Options{ProtectedKeys: []string{"DB_PASSWORD"}}
	out := classifier.Classify(results, opts)
	if out[0].Level != classifier.LevelCritical {
		t.Errorf("expected critical, got %s", out[0].Level)
	}
}

func TestSortByLevel_OrdersCorrectly(t *testing.T) {
	input := []classifier.ClassifiedResult{
		{Result: makeResult("svc-a", "k1"), Level: classifier.LevelLow},
		{Result: makeResult("svc-b"), Level: classifier.LevelNone},
		{Result: makeResult("svc-c", "k1", "k2", "k3"), Level: classifier.LevelMedium},
		{Result: makeResult("svc-d", "secret"), Level: classifier.LevelCritical},
	}
	sorted := classifier.SortByLevel(input)
	expected := []classifier.Level{
		classifier.LevelCritical,
		classifier.LevelMedium,
		classifier.LevelLow,
		classifier.LevelNone,
	}
	for i, r := range sorted {
		if r.Level != expected[i] {
			t.Errorf("index %d: expected %s, got %s", i, expected[i], r.Level)
		}
	}
}

func TestClassify_CustomThresholds(t *testing.T) {
	opts := classifier.Options{LowThreshold: 1, MediumThreshold: 2, HighThreshold: 4}
	results := []drift.Result{makeResult("svc-a", "k1", "k2")}
	out := classifier.Classify(results, opts)
	if out[0].Level != classifier.LevelMedium {
		t.Errorf("expected medium with custom threshold, got %s", out[0].Level)
	}
}

package censor_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/censor"
	"github.com/driftwatch/internal/drift"
)

func makeResult(service string, diffs ...drift.Diff) drift.Result {
	return drift.Result{
		Service:   service,
		Drifted:   len(diffs) > 0,
		Diffs:     diffs,
		ScannedAt: time.Now(),
	}
}

func diff(key, expected, actual string) drift.Diff {
	return drift.Diff{Key: key, Expected: expected, Actual: actual}
}

func TestApply_NoPatterns_Unchanged(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", diff("password", "secret", "other")),
	}
	out := censor.Apply(results, censor.Options{})
	if out[0].Diffs[0].Expected != "secret" {
		t.Errorf("expected value unchanged, got %q", out[0].Diffs[0].Expected)
	}
}

func TestApply_CensorsMatchingKey(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", diff("db_password", "secret", "wrong")),
	}
	out := censor.Apply(results, censor.Options{Patterns: []string{"password"}})
	d := out[0].Diffs[0]
	if d.Expected != "***" || d.Actual != "***" {
		t.Errorf("expected censored values, got expected=%q actual=%q", d.Expected, d.Actual)
	}
}

func TestApply_CaseInsensitiveMatch(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-b", diff("API_SECRET", "tok", "bad")),
	}
	out := censor.Apply(results, censor.Options{Patterns: []string{"secret"}})
	if out[0].Diffs[0].Actual != "***" {
		t.Error("expected case-insensitive match to be censored")
	}
}

func TestApply_CustomReplacement(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-c", diff("token", "abc", "xyz")),
	}
	out := censor.Apply(results, censor.Options{Patterns: []string{"token"}, Replacement: "[REDACTED]"})
	if out[0].Diffs[0].Expected != "[REDACTED]" {
		t.Errorf("unexpected replacement: %q", out[0].Diffs[0].Expected)
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	original := []drift.Result{
		makeResult("svc-d", diff("secret_key", "real", "fake")),
	}
	censor.Apply(original, censor.Options{Patterns: []string{"secret"}})
	if original[0].Diffs[0].Expected != "real" {
		t.Error("original result was mutated")
	}
}

func TestApply_MultiplePatterns(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-e",
			diff("api_key", "k1", "k2"),
			diff("db_pass", "p1", "p2"),
			diff("log_level", "info", "debug"),
		),
	}
	out := censor.Apply(results, censor.Options{Patterns: []string{"api_key", "pass"}})
	if out[0].Diffs[0].Actual != "***" {
		t.Error("api_key should be censored")
	}
	if out[0].Diffs[1].Actual != "***" {
		t.Error("db_pass should be censored")
	}
	if out[0].Diffs[2].Actual != "debug" {
		t.Error("log_level should not be censored")
	}
}

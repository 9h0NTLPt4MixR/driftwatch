package redactor_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/redactor"
)

func makeResult(service string, expected, actual map[string]string, drifted bool) drift.DriftResult {
	var diffs []drift.Diff
	if drifted {
		for k := range actual {
			if expected[k] != actual[k] {
				diffs = append(diffs, drift.Diff{Key: k, Expected: expected[k], Actual: actual[k]})
			}
		}
	}
	return drift.DriftResult{
		Service:  service,
		Expected: expected,
		Actual:   actual,
		Diffs:    diffs,
		HasDrift: drifted,
	}
}

func TestApply_NoSensitiveKeys(t *testing.T) {
	results := []drift.DriftResult{
		makeResult("svc", map[string]string{"host": "localhost"}, map[string]string{"host": "localhost"}, false),
	}
	out := redactor.Apply(results, redactor.Options{})
	if out[0].Expected["host"] != "localhost" {
		t.Errorf("expected 'localhost', got %q", out[0].Expected["host"])
	}
}

func TestApply_RedactsPassword(t *testing.T) {
	results := []drift.DriftResult{
		makeResult("svc",
			map[string]string{"db_password": "secret123", "host": "db"},
			map[string]string{"db_password": "changed", "host": "db"},
			true),
	}
	out := redactor.Apply(results, redactor.Options{})
	if out[0].Expected["db_password"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %q", out[0].Expected["db_password"])
	}
	if out[0].Actual["db_password"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %q", out[0].Actual["db_password"])
	}
	if out[0].Expected["host"] != "db" {
		t.Error("non-sensitive key should not be redacted")
	}
}

func TestApply_RedactsDiffs(t *testing.T) {
	results := []drift.DriftResult{
		{
			Service:  "svc",
			HasDrift: true,
			Diffs: []drift.Diff{
				{Key: "api_token", Expected: "old", Actual: "new"},
				{Key: "region", Expected: "us-east", Actual: "eu-west"},
			},
		},
	}
	out := redactor.Apply(results, redactor.Options{})
	for _, d := range out[0].Diffs {
		if d.Key == "api_token" {
			if d.Expected != "[REDACTED]" || d.Actual != "[REDACTED]" {
				t.Errorf("api_token diff should be redacted, got expected=%q actual=%q", d.Expected, d.Actual)
			}
		}
		if d.Key == "region" && d.Expected == "[REDACTED]" {
			t.Error("region should not be redacted")
		}
	}
}

func TestApply_CustomSensitiveKeys(t *testing.T) {
	results := []drift.DriftResult{
		makeResult("svc",
			map[string]string{"internal_cert": "abc", "host": "x"},
			map[string]string{"internal_cert": "abc", "host": "x"},
			false),
	}
	out := redactor.Apply(results, redactor.Options{SensitiveKeys: []string{"cert"}})
	if out[0].Expected["internal_cert"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED] for custom key, got %q", out[0].Expected["internal_cert"])
	}
	if out[0].Expected["host"] == "[REDACTED]" {
		t.Error("host should not be redacted with custom keys")
	}
}

func TestApply_NilMaps(t *testing.T) {
	results := []drift.DriftResult{{Service: "empty"}}
	out := redactor.Apply(results, redactor.Options{})
	if out[0].Expected != nil || out[0].Actual != nil {
		t.Error("nil maps should remain nil after redaction")
	}
}

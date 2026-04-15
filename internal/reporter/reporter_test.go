package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/driftwatch/internal/drift"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{Service: service, Diffs: diffs}
}

func TestWriteText_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	results := []drift.Result{makeResult("svc-a", nil)}
	if err := Write(&buf, results, FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "[OK]") {
		t.Errorf("expected [OK] in output, got: %s", buf.String())
	}
}

func TestWriteText_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	diffs := []drift.Diff{{Key: "version", Expected: "1.0", Actual: "2.0"}}
	results := []drift.Result{makeResult("svc-b", diffs)}
	if err := Write(&buf, results, FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[DRIFT]") {
		t.Errorf("expected [DRIFT] in output")
	}
	if !strings.Contains(out, "version") {
		t.Errorf("expected key 'version' in output")
	}
}

func TestWriteJSON_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	results := []drift.Result{makeResult("svc-c", nil)}
	if err := Write(&buf, results, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "svc-c") {
		t.Errorf("expected service name in JSON output")
	}
	if !strings.Contains(out, "\"drift\": false") {
		t.Errorf("expected drift:false in JSON output, got: %s", out)
	}
}

func TestWriteJSON_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	diffs := []drift.Diff{{Key: "env", Expected: "prod", Actual: "staging"}}
	results := []drift.Result{makeResult("svc-d", diffs)}
	if err := Write(&buf, results, FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "\"drift\": true") {
		t.Errorf("expected drift:true in JSON output, got: %s", out)
	}
	if !strings.Contains(out, "env") {
		t.Errorf("expected key 'env' in JSON output")
	}
}

func TestWrite_DefaultsToText(t *testing.T) {
	var buf bytes.Buffer
	results := []drift.Result{makeResult("svc-e", nil)}
	if err := Write(&buf, results, "unknown"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "[OK]") {
		t.Errorf("expected text format fallback")
	}
}

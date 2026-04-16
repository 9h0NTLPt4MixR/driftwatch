package notifier_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/user/driftwatch/internal/drift"
	"github.com/user/driftwatch/internal/notifier"
)

func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func makeResults(hasDrift bool) []drift.Result {
	diffs := []drift.Diff{}
	if hasDrift {
		diffs = append(diffs, drift.Diff{Key: "replicas", Expected: "3", Actual: "1"})
	}
	return []drift.Result{
		{Service: "api", HasDrift: hasDrift, Diffs: diffs},
	}
}

func TestNotify_NoDrift(t *testing.T) {
	out := captureStdout(func() {
		notifier.Notify(makeResults(false), notifier.Options{Channel: notifier.ChannelStdout})
	})
	if !strings.Contains(out, "0 with drift") {
		t.Errorf("expected 0 drift summary, got: %s", out)
	}
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK status, got: %s", out)
	}
}

func TestNotify_WithDrift(t *testing.T) {
	out := captureStdout(func() {
		notifier.Notify(makeResults(true), notifier.Options{Channel: notifier.ChannelStdout})
	})
	if !strings.Contains(out, "1 with drift") {
		t.Errorf("expected 1 drift, got: %s", out)
	}
	if !strings.Contains(out, "DRIFT") {
		t.Errorf("expected DRIFT status, got: %s", out)
	}
}

func TestNotify_OnlyDrift_SkipsClean(t *testing.T) {
	out := captureStdout(func() {
		notifier.Notify(makeResults(false), notifier.Options{
			Channel:   notifier.ChannelStdout,
			OnlyDrift: true,
		})
	})
	if out != "" {
		t.Errorf("expected no output for clean results with OnlyDrift, got: %s", out)
	}
}

func TestNotify_OnlyDrift_PrintsDrift(t *testing.T) {
	out := captureStdout(func() {
		notifier.Notify(makeResults(true), notifier.Options{
			Channel:   notifier.ChannelStdout,
			OnlyDrift: true,
		})
	})
	if !strings.Contains(out, "DRIFT") {
		t.Errorf("expected DRIFT in output, got: %s", out)
	}
	_ = fmt.Sprintf("%s", out)
}

package exporter_test

import (
	"strings"
	"testing"

	"github.com/user/driftwatch/internal/drift"
	"github.com/user/driftwatch/internal/exporter"
)

func makeResults() []drift.Result {
	return []drift.Result{
		{
			Service:  "api",
			HasDrift: true,
			Diffs: []drift.Diff{
				{Key: "LOG_LEVEL", Expected: "info", Actual: "debug"},
			},
		},
		{
			Service:  "worker",
			HasDrift: false,
			Diffs:    nil,
		},
	}
}

func TestWrite_CSV(t *testing.T) {
	var sb strings.Builder
	if err := exporter.Write(&sb, makeResults(), exporter.FormatCSV); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "service,key,expected,actual") {
		t.Error("missing CSV header")
	}
	if !strings.Contains(out, "api,LOG_LEVEL,info,debug") {
		t.Error("missing drift row")
	}
	if !strings.Contains(out, "worker,,") {
		t.Error("missing clean service row")
	}
}

func TestWrite_Markdown(t *testing.T) {
	var sb strings.Builder
	if err := exporter.Write(&sb, makeResults(), exporter.FormatMarkdown); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "| Service |") {
		t.Error("missing markdown header")
	}
	if !strings.Contains(out, "api") || !strings.Contains(out, "LOG_LEVEL") {
		t.Error("missing drift data in markdown")
	}
	if !strings.Contains(out, "worker") {
		t.Error("missing clean service in markdown")
	}
}

func TestWrite_UnsupportedFormat(t *testing.T) {
	var sb strings.Builder
	err := exporter.Write(&sb, makeResults(), exporter.Format("xml"))
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

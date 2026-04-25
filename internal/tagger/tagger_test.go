package tagger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/tagger"
)

func makeResult(service string, drifted bool) drift.Result {
	r := drift.Result{Service: service}
	if drifted {
		r.Diffs = []drift.Diff{{Key: "timeout", Expected: "30s", Actual: "60s"}}
	}
	return r
}

func TestApply_NoTags(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", false)}
	out := tagger.Apply(results, tagger.TagMap{})
	if len(out[0].Tags) != 0 {
		t.Fatalf("expected no tags, got %v", out[0].Tags)
	}
}

func TestApply_AssignsTags(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", true), makeResult("svc-b", false)}
	tm := tagger.TagMap{
		"svc-a": {"critical", "prod"},
	}
	out := tagger.Apply(results, tm)
	if len(out[0].Tags) != 2 {
		t.Fatalf("expected 2 tags for svc-a, got %v", out[0].Tags)
	}
	if len(out[1].Tags) != 0 {
		t.Fatalf("expected no tags for svc-b, got %v", out[1].Tags)
	}
}

func TestApply_DedupesTags(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", false)}
	tm := tagger.TagMap{"svc-a": {"prod", "prod", "critical"}}
	out := tagger.Apply(results, tm)
	if len(out[0].Tags) != 2 {
		t.Fatalf("expected 2 unique tags, got %v", out[0].Tags)
	}
}

func TestFilterByTag_ReturnsMatches(t *testing.T) {
	results := []drift.Result{
		{Service: "svc-a", Tags: []string{"critical"}},
		{Service: "svc-b", Tags: []string{"staging"}},
		{Service: "svc-c", Tags: []string{"critical", "prod"}},
	}
	out := tagger.FilterByTag(results, "critical")
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestFilterByTag_NoMatch(t *testing.T) {
	results := []drift.Result{{Service: "svc-a", Tags: []string{"staging"}}}
	out := tagger.FilterByTag(results, "prod")
	if len(out) != 0 {
		t.Fatalf("expected 0 results, got %d", len(out))
	}
}

func TestLoadTagMap_Valid(t *testing.T) {
	tm := tagger.TagMap{"svc-a": {"prod"}, "svc-b": {"staging"}}
	data, _ := json.Marshal(tm)
	f, _ := os.CreateTemp(t.TempDir(), "tags-*.json")
	_ = os.WriteFile(f.Name(), data, 0644)

	loaded, err := tagger.LoadTagMap(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded["svc-a"]) != 1 || loaded["svc-a"][0] != "prod" {
		t.Fatalf("unexpected tag map: %v", loaded)
	}
}

func TestLoadTagMap_MissingFile(t *testing.T) {
	_, err := tagger.LoadTagMap(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

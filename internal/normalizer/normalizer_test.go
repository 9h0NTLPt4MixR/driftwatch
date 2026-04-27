package normalizer_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/normalizer"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service:   service,
		Endpoint:  "http://example.com",
		Drifted:   len(diffs) > 0,
		ScannedAt: time.Now(),
		Diffs:     diffs,
	}
}

func TestApply_NoOptions_Unchanged(t *testing.T) {
	results := []drift.Result{
		makeResult("svc", []drift.Diff{{Key: "PORT", Want: " 8080 ", Got: " 9090 "}}),
	}
	out := normalizer.Apply(results, normalizer.Options{})
	if out[0].Diffs[0].Want != " 8080 " {
		t.Errorf("expected value unchanged, got %q", out[0].Diffs[0].Want)
	}
}

func TestApply_TrimSpace(t *testing.T) {
	results := []drift.Result{
		makeResult("svc", []drift.Diff{{Key: "HOST", Want: "  localhost  ", Got: " remotehost "}})}
	out := normalizer.Apply(results, normalizer.Options{TrimSpace: true})
	if out[0].Diffs[0].Want != "localhost" {
		t.Errorf("expected trimmed want, got %q", out[0].Diffs[0].Want)
	}
	if out[0].Diffs[0].Got != "remotehost" {
		t.Errorf("expected trimmed got, got %q", out[0].Diffs[0].Got)
	}
}

func TestApply_LowercaseKeys(t *testing.T) {
	results := []drift.Result{
		makeResult("svc", []drift.Diff{{Key: "DB_HOST", Want: "a", Got: "b"}}),
	}
	out := normalizer.Apply(results, normalizer.Options{LowercaseKeys: true})
	if out[0].Diffs[0].Key != "db_host" {
		t.Errorf("expected lowercase key, got %q", out[0].Diffs[0].Key)
	}
}

func TestApply_CoerceBooleans(t *testing.T) {
	cases := []struct{ input, want string }{
		{"1", "true"}, {"yes", "true"}, {"TRUE", "true"},
		{"0", "false"}, {"no", "false"}, {"FALSE", "false"},
	}
	for _, tc := range cases {
		results := []drift.Result{
			makeResult("svc", []drift.Diff{{Key: "flag", Want: tc.input, Got: tc.input}}),
		}
		out := normalizer.Apply(results, normalizer.Options{CoerceBooleans: true})
		if out[0].Diffs[0].Want != tc.want {
			t.Errorf("CoerceBooleans(%q): expected %q, got %q", tc.input, tc.want, out[0].Diffs[0].Want)
		}
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	orig := makeResult("svc", []drift.Diff{{Key: "KEY", Want: "  val  ", Got: "  other  "}})
	results := []drift.Result{orig}
	normalizer.Apply(results, normalizer.Options{TrimSpace: true})
	if results[0].Diffs[0].Want != "  val  " {
		t.Error("Apply must not mutate original results")
	}
}

func TestApply_EmptyResults(t *testing.T) {
	out := normalizer.Apply([]drift.Result{}, normalizer.Options{TrimSpace: true})
	if len(out) != 0 {
		t.Errorf("expected empty output, got %d results", len(out))
	}
}

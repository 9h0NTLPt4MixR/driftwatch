package ranker_test

import (
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/ranker"
)

func makeResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{
		Service:  service,
		HasDrift: len(diffs) > 0,
		Diffs:    diffs,
	}
}

func TestRank_NoDrift(t *testing.T) {
	results := []drift.Result{makeResult("svc-a", nil)}
	ranked := ranker.Rank(results, ranker.Options{})
	if len(ranked) != 1 {
		t.Fatalf("expected 1 result, got %d", len(ranked))
	}
	if ranked[0].Score != 0 || ranked[0].Severity != "clean" {
		t.Errorf("expected clean score, got score=%d severity=%s", ranked[0].Score, ranked[0].Severity)
	}
}

func TestRank_SortsDescending(t *testing.T) {
	results := []drift.Result{
		makeResult("low", []drift.Diff{{Key: "a", Expected: "1", Actual: "2"}}),
		makeResult("high", []drift.Diff{
			{Key: "a"}, {Key: "b"}, {Key: "c"}, {Key: "d"}, {Key: "e"},
		}),
		makeResult("clean", nil),
	}
	ranked := ranker.Rank(results, ranker.Options{})
	if ranked[0].Service != "high" {
		t.Errorf("expected 'high' first, got %s", ranked[0].Service)
	}
	if ranked[len(ranked)-1].Service != "clean" {
		t.Errorf("expected 'clean' last, got %s", ranked[len(ranked)-1].Service)
	}
}

func TestRank_ProtectedKeyWeight(t *testing.T) {
	results := []drift.Result{
		makeResult("svc-a", []drift.Diff{{Key: "secret", Expected: "x", Actual: "y"}}),
		makeResult("svc-b", []drift.Diff{
			{Key: "a"}, {Key: "b"}, {Key: "c"},
		}),
	}
	opts := ranker.Options{ProtectedKeys: []string{"secret"}, ProtectedWeight: 5}
	ranked := ranker.Rank(results, opts)
	// svc-a: 1 protected key * 5 = 5; svc-b: 3 normal = 3
	if ranked[0].Service != "svc-a" {
		t.Errorf("expected svc-a first due to protected weight, got %s", ranked[0].Service)
	}
	if ranked[0].Score != 5 {
		t.Errorf("expected score 5, got %d", ranked[0].Score)
	}
}

func TestRank_SeverityLabels(t *testing.T) {
	cases := []struct {
		diffs    int
		want     string
	}{
		{0, "clean"},
		{1, "low"},
		{3, "low"},
		{4, "medium"},
		{7, "medium"},
		{8, "high"},
		{15, "critical"},
	}
	for _, tc := range cases {
		diffs := make([]drift.Diff, tc.diffs)
		for i := range diffs {
			diffs[i] = drift.Diff{Key: "k"}
		}
		ranked := ranker.Rank([]drift.Result{makeResult("svc", diffs)}, ranker.Options{})
		if ranked[0].Severity != tc.want {
			t.Errorf("diffs=%d: expected severity %q, got %q", tc.diffs, tc.want, ranked[0].Severity)
		}
	}
}

func TestRank_DefaultProtectedWeight(t *testing.T) {
	results := []drift.Result{
		makeResult("svc", []drift.Diff{{Key: "password"}}),
	}
	opts := ranker.Options{ProtectedKeys: []string{"password"}, ProtectedWeight: 0}
	ranked := ranker.Rank(results, opts)
	// default weight 3
	if ranked[0].Score != 3 {
		t.Errorf("expected default weight score 3, got %d", ranked[0].Score)
	}
}

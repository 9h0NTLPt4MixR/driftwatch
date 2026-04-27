package sampler_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/sampler"
)

func makeResult(service string, drifted bool) drift.Result {
	return drift.Result{
		Service:   service,
		Drifted:   drifted,
		ScannedAt: time.Now(),
	}
}

func makeResults(n int) []drift.Result {
	out := make([]drift.Result, n)
	for i := 0; i < n; i++ {
		out[i] = makeResult(fmt.Sprintf("svc-%d", i), i%2 == 0)
	}
	return out
}

func TestSample_InvalidRate(t *testing.T) {
	_, err := sampler.Sample(nil, sampler.Options{Rate: 1.5})
	if err == nil {
		t.Fatal("expected error for rate > 1")
	}
	_, err = sampler.Sample(nil, sampler.Options{Rate: -0.1})
	if err == nil {
		t.Fatal("expected error for rate < 0")
	}
}

func TestSample_EmptyInput(t *testing.T) {
	got, err := sampler.Sample([]drift.Result{}, sampler.Options{Rate: 0.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

func TestSample_Head_HalfRate(t *testing.T) {
	results := makeResults(10)
	got, err := sampler.Sample(results, sampler.Options{Rate: 0.5, Strategy: sampler.StrategyHead})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5, got %d", len(got))
	}
}

func TestSample_Head_FullRate(t *testing.T) {
	results := makeResults(8)
	got, _ := sampler.Sample(results, sampler.Options{Rate: 1.0, Strategy: sampler.StrategyHead})
	if len(got) != 8 {
		t.Fatalf("expected 8, got %d", len(got))
	}
}

func TestSample_Modulo_HalfRate(t *testing.T) {
	results := makeResults(10)
	got, err := sampler.Sample(results, sampler.Options{Rate: 0.5, Strategy: sampler.StrategyModulo})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("expected 5, got %d", len(got))
	}
}

func TestSample_Random_Reproducible(t *testing.T) {
	results := makeResults(100)
	opts := sampler.Options{Rate: 0.5, Strategy: sampler.StrategyRandom, Seed: 42}
	a, _ := sampler.Sample(results, opts)
	b, _ := sampler.Sample(results, opts)
	if len(a) != len(b) {
		t.Fatalf("expected same count: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Service != b[i].Service {
			t.Fatalf("mismatch at index %d", i)
		}
	}
}

func TestHashSample_Deterministic(t *testing.T) {
	results := makeResults(50)
	a, err := sampler.HashSample(results, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, _ := sampler.HashSample(results, 0.5)
	if len(a) != len(b) {
		t.Fatalf("expected same count: %d vs %d", len(a), len(b))
	}
}

func TestHashSample_InvalidRate(t *testing.T) {
	_, err := sampler.HashSample(nil, 2.0)
	if err == nil {
		t.Fatal("expected error")
	}
}

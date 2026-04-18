package scorecard_test

import (
	"testing"

	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/scorecard"
)

func BenchmarkCompute(b *testing.B) {
	results := make([]drift.Result, 200)
	for i := range results {
		hasDrift := i%3 == 0
		var diffs []drift.Diff
		if hasDrift {
			diffs = []drift.Diff{
				{Key: "timeout", Expected: "30s", Actual: "60s"},
				{Key: "replicas", Expected: "2", Actual: "1"},
			}
		}
		results[i] = drift.Result{
			Service:  fmt.Sprintf("svc-%d", i),
			HasDrift: hasDrift,
			Diffs:    diffs,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scorecard.Compute(results)
	}
}

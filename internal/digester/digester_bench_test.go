package digester_test

import (
	"fmt"
	"testing"

	"github.com/driftwatch/internal/digester"
	"github.com/driftwatch/internal/drift"
)

func BenchmarkCompute(b *testing.B) {
	results := make([]drift.Result, 50)
	for i := range results {
		diffs := make([]drift.Diff, 10)
		for j := range diffs {
			diffs[j] = drift.Diff{
				Key:      fmt.Sprintf("key_%d", j),
				Expected: fmt.Sprintf("expected_%d", j),
				Actual:   fmt.Sprintf("actual_%d", j),
			}
		}
		results[i] = makeResult(fmt.Sprintf("service_%d", i), diffs)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = digester.Compute(results)
	}
}

func BenchmarkChanged(b *testing.B) {
	results := make([]drift.Result, 50)
	for i := range results {
		results[i] = makeResult(fmt.Sprintf("service_%d", i), nil)
	}
	prev, _ := digester.Compute(results)
	next, _ := digester.Compute(results)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = digester.Changed(prev, next)
	}
}

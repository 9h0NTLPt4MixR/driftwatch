package labeler_test

import (
	"fmt"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/labeler"
)

func BenchmarkApply(b *testing.B) {
	const numResults = 500
	const numRules = 20

	results := make([]drift.Result, numResults)
	for i := range results {
		results[i] = drift.Result{
			Service: fmt.Sprintf("svc-%d", i),
			Drifted: i%3 == 0,
			Diffs: []drift.Diff{
				{Key: fmt.Sprintf("auth.key%d", i), Expected: "old", Actual: "new"},
			},
		}
	}

	rules := make([]labeler.Rule, numRules)
	for i := range rules {
		rules[i] = labeler.Rule{
			Label:     fmt.Sprintf("label-%d", i),
			KeyPrefix: fmt.Sprintf("auth."),
			MinDrifted: 1,
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = labeler.Apply(results, rules)
	}
}

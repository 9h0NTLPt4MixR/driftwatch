package policy_test

import (
	"fmt"
	"testing"

	"github.com/user/driftwatch/internal/drift"
	"github.com/user/driftwatch/internal/policy"
)

func BenchmarkEvaluate(b *testing.B) {
	const numServices = 50
	const numKeys = 20

	results := make([]drift.Result, numServices)
	for i := range results {
		expected := make(map[string]string, numKeys)
		var diffs []drift.Diff
		for k := 0; k < numKeys; k++ {
			key := fmt.Sprintf("KEY_%d", k)
			expected[key] = "value"
			if k%5 == 0 {
				diffs = append(diffs, drift.Diff{Key: key, Expected: "value", Actual: "changed"})
			}
		}
		results[i] = drift.Result{
			Service:  fmt.Sprintf("service-%d", i),
			Expected: expected,
			Diffs:    diffs,
		}
	}

	rules := []policy.Rule{
		{Name: "protected", ProtectedKeys: []string{"KEY_0", "KEY_10"}, MaxDriftPercent: 100},
		{Name: "threshold", MaxDriftPercent: 15},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy.Evaluate(results, rules)
	}
}

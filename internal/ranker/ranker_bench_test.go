package ranker_test

import (
	"fmt"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/ranker"
)

func BenchmarkRank(b *testing.B) {
	const numServices = 200
	const diffsPerService = 10

	results := make([]drift.Result, numServices)
	for i := 0; i < numServices; i++ {
		diffs := make([]drift.Diff, diffsPerService)
		for j := 0; j < diffsPerService; j++ {
			diffs[j] = drift.Diff{
				Key:      fmt.Sprintf("key_%d", j),
				Expected: "old",
				Actual:   "new",
			}
		}
		results[i] = drift.Result{
			Service:  fmt.Sprintf("service-%d", i),
			HasDrift: true,
			Diffs:    diffs,
		}
	}

	opts := ranker.Options{
		ProtectedKeys:   []string{"key_0", "key_1", "key_2"},
		ProtectedWeight: 3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ranker.Rank(results, opts)
	}
}

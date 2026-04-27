package enricher_test

import (
	"fmt"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/enricher"
)

func BenchmarkApply(b *testing.B) {
	const n = 500
	results := make([]drift.Result, n)
	meta := make(map[string]enricher.Meta, n)
	for i := 0; i < n; i++ {
		svc := fmt.Sprintf("service-%d", i)
		results[i] = makeResult(svc, i%3 == 0)
		meta[svc] = enricher.Meta{
			Owner:       "team-x",
			Tier:        "standard",
			Environment: "production",
		}
	}
	opts := enricher.Options{
		DefaultEnvironment: "production",
		ServiceMeta:        meta,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = enricher.Apply(results, opts)
	}
}

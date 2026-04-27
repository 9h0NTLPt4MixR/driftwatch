package indexer_test

import (
	"fmt"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/indexer"
)

func BenchmarkBuild(b *testing.B) {
	const n = 500
	results := make([]drift.Result, n)
	for i := 0; i < n; i++ {
		results[i] = makeResult(
			fmt.Sprintf("service-%d", i),
			i%2 == 0,
			"timeout", "retries", "log_level",
		)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = indexer.Build(results)
	}
}

func BenchmarkLookupService(b *testing.B) {
	const n = 500
	results := make([]drift.Result, n)
	for i := 0; i < n; i++ {
		results[i] = makeResult(fmt.Sprintf("service-%d", i), i%2 == 0)
	}
	idx, _ := indexer.Build(results)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ , _ = idx.LookupService(fmt.Sprintf("service-%d", i%n))
	}
}

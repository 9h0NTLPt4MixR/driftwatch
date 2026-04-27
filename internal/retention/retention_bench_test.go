package retention_test

import (
	"testing"
	"time"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/retention"
)

func BenchmarkApply(b *testing.B) {
	results := make([]drift.Result, 500)
	for i := range results {
		results[i] = drift.Result{
			Service:   "svc",
			HasDrift:  i%3 == 0,
			Timestamp: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}
	p := retention.Policy{
		MaxAge:      6 * time.Hour,
		MaxEntries:  200,
		KeepDrifted: true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = retention.Apply(results, p)
	}
}

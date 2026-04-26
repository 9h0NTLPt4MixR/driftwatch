// Package rollup provides time-bucketed aggregation of drift scan results.
//
// It groups [drift.Result] values into hourly, daily, or weekly windows
// and computes per-bucket statistics including drift rate, total diffs,
// and counts of drifted vs clean services.
//
// Typical usage:
//
//	buckets := rollup.Compute(results, rollup.Daily)
//	for _, b := range buckets {
//		fmt.Printf("%s drift_rate=%.2f\n", b.WindowStart.Format(time.DateOnly), b.DriftRate)
//	}
package rollup

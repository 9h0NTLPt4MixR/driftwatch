// Package retention implements age- and count-based retention policies for
// drift scan results.
//
// Use Apply to filter a slice of drift.Result values according to a Policy.
// The Policy supports:
//
//   - MaxAge: discard results older than a given duration.
//   - MaxEntries: keep only the N most-recent results.
//   - KeepDrifted: exempt drifted results from age-based removal.
//
// Example:
//
//	p := retention.Policy{MaxAge: 7 * 24 * time.Hour, MaxEntries: 100}
//	out, err := retention.Apply(results, p)
package retention

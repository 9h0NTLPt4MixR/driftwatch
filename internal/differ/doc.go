// Package differ computes and formats configuration diffs between the
// expected (declared) state and the actual (live) state of a service.
//
// It consumes drift.Result values produced by the drift package and
// converts them into structured Diff objects that can be rendered as
// human-readable unified-style output or consumed programmatically.
//
// Typical usage:
//
//	results := drift.Scan(cfg, fetcher)
//	diffs   := differ.Compute(results)
//	fmt.Print(differ.Format(diffs))
package differ
